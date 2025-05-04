package services

import (
	"errors"
	"fmt"
	"time"

	"bankapp/internal/config"
	"bankapp/internal/models"
	"bankapp/internal/repo"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// содержит все репозитории и конфиг
type BankService struct {
	userRepo        *repo.UserRepo
	accountRepo     *repo.AccountRepo
	cardRepo        *repo.CardRepo
	transactionRepo *repo.TransactionRepo
	creditRepo      *repo.CreditRepo
	scheduleRepo    *repo.ScheduleRepo
	cfg             *config.Config
}

// конструктор
func NewBankService(
	u *repo.UserRepo,
	a *repo.AccountRepo,
	c *repo.CardRepo,
	t *repo.TransactionRepo,
	cr *repo.CreditRepo,
	s *repo.ScheduleRepo,
	cfg *config.Config,
) *BankService {
	return &BankService{u, a, c, t, cr, s, cfg}
}

// регистрация нового пользователя
func (s *BankService) RegisterUser(req models.RegisterRequest) (*models.User, error) {
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
	}
	if err := user.Validate(req.Password); err != nil {
		return nil, err
	}
	// проверка уникальности
	if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
		return nil, fmt.Errorf("имя %s уже занято", req.Username)
	}
	if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
		return nil, fmt.Errorf("email %s уже зарегистрирован", req.Email)
	}
	// хешируем пароль
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = hash
	// сохраняем в БД
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	// асинхронно шлём welcome-письмо
	go sendEmailNotification(
		s.cfg,
		user.Email,
		"Добро пожаловать в BankApp",
		fmt.Sprintf("Здравствуйте, %s! Спасибо за регистрацию.", user.Username),
	)
	// не возвращаем хеш обратно
	user.PasswordHash = ""
	return user, nil
}

// JWT
func (s *BankService) LoginUser(req models.LoginRequest) (string, error) {
	u, err := s.userRepo.GetByUsername(req.Username)
	if err != nil || !CheckPasswordHash(req.Password, u.PasswordHash) {
		return "", errors.New("неверное имя или пароль")
	}
	return s.generateJWT(u.ID.String())
}

func (s *BankService) ParseToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("некорректный токен")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("неверные claims")
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("нет subject в токене")
	}
	return uuid.Parse(sub)
}

// создаёт JWT с 24-часовым сроком
func (s *BankService) generateJWT(userID string) (string, error) {
	exp := time.Now().Add(24 * time.Hour).Unix()
	claims := jwt.MapClaims{"sub": userID, "exp": exp}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// новый счёт для userID
func (s *BankService) CreateAccount(userID uuid.UUID) (*models.Account, error) {
	acc := &models.Account{
		UserID:  userID,
		Number:  generateAccountNumber(),
		Balance: decimal.Zero,
	}
	if err := s.accountRepo.Create(acc); err != nil {
		return nil, err
	}
	return acc, nil
}

// список счетов пользователя
func (s *BankService) GetUserAccounts(userID uuid.UUID) ([]models.Account, error) {
	return s.accountRepo.GetByUserID(userID)
}

// карта к счёту
func (s *BankService) GenerateCard(accountID uuid.UUID) (*models.Card, error) {
	if _, err := s.accountRepo.GetByID(accountID); err != nil {
		return nil, fmt.Errorf("счёт %s не найден", accountID)
	}
	cardNum := generateCardNumber()
	expM, expY := generateExpiryDate()
	cvv := generateCVV()

	numEnc, err := EncryptPGP([]byte(cardNum), s.cfg.PGPPublicKeyPath)
	if err != nil {
		return nil, err
	}
	expiryStr := fmt.Sprintf("%02d/%02d", expM, expY%100)
	expEnc, err := EncryptPGP([]byte(expiryStr), s.cfg.PGPPublicKeyPath)
	if err != nil {
		return nil, err
	}
	cvvHash, err := HashCVV(cvv)
	if err != nil {
		return nil, err
	}
	hmacHex := ComputeHMAC(cardNum+expiryStr, []byte(s.cfg.HMACSecret))

	card := &models.Card{
		AccountID: accountID,
		NumberEnc: numEnc,
		ExpiryEnc: expEnc,
		CVVHash:   cvvHash,
		HMAC:      hmacHex,
	}
	if err := s.cardRepo.Create(card); err != nil {
		return nil, err
	}
	card.CVVHash = "***"
	return card, nil
}

// список карт по счёту
func (s *BankService) GetAccountCards(accountID uuid.UUID) ([]models.Card, error) {
	cards, err := s.cardRepo.GetByAccountID(accountID)
	if err != nil {
		return nil, err
	}
	for i := range cards {
		cards[i].CVVHash = "***"
	}
	return cards, nil
}

// оплата по карте
func (s *BankService) PayWithCard(req models.PaymentRequest) error {
	if req.Amount.LessOrEqual(decimal.Zero) {
		return errors.New("сумма должна быть >0")
	}
	hmacHex := ComputeHMAC(req.CardNumber, []byte(s.cfg.HMACSecret))
	card, err := s.cardRepo.GetByHMAC(hmacHex)
	if err != nil {
		return errors.New("карта не найдена")
	}
	// запуск транзакции
	return s.accountRepo.WithTx(func(tx repo.TxContext) error {
		acc, err := s.accountRepo.GetByID(card.AccountID)
		if err != nil {
			return err
		}
		if acc.Balance.LessThan(req.Amount) {
			return errors.New("недостаточно средств")
		}
		newBal := acc.Balance.Sub(req.Amount)
		if err := s.accountRepo.UpdateBalanceTx(tx, acc.ID, newBal); err != nil {
			return err
		}
		tr := &models.Transaction{
			From:      &acc.ID,
			To:        nil,
			Amount:    req.Amount,
			Type:      "payment",
			Note:      fmt.Sprintf("оплата %s", req.Merchant),
			CreatedAt: time.Now(),
		}
		if err := s.transactionRepo.CreateTx(tx, tr); err != nil {
			return err
		}
		// уведомление
		go func() {
			u, _ := s.userRepo.GetByID(acc.UserID)
			_ = sendEmailNotification(
				s.cfg,
				u.Email,
				"Успешная оплата",
				fmt.Sprintf("Вы оплатили %s: %s", req.Merchant, req.Amount),
			)
		}()
		return nil
	})
}

// перевод между счетами
func (s *BankService) Transfer(req models.TransferRequest) error {
	if req.Amount.LessOrEqual(decimal.Zero) {
		return errors.New("сумма должна быть >0")
	}
	if req.FromAccountID == req.ToAccountID {
		return errors.New("невозможно перевести на тот же счёт")
	}
	return s.accountRepo.WithTx(func(tx repo.TxContext) error {
		fromAcc, err := s.accountRepo.GetByID(req.FromAccountID)
		if err != nil {
			return err
		}
		toAcc, err := s.accountRepo.GetByID(req.ToAccountID)
		if err != nil {
			return err
		}
		if fromAcc.Balance.LessThan(req.Amount) {
			return errors.New("недостаточно средств")
		}
		// списываем со счёта отправителя
		if err := s.accountRepo.UpdateBalanceTx(tx, fromAcc.ID, fromAcc.Balance.Sub(req.Amount)); err != nil {
			return err
		}
		// зачисляем на счёт получателя
		if err := s.accountRepo.UpdateBalanceTx(tx, toAcc.ID, toAcc.Balance.Add(req.Amount)); err != nil {
			return err
		}
		tr := &models.Transaction{
			From:      &fromAcc.ID,
			To:        &toAcc.ID,
			Amount:    req.Amount,
			Type:      "transfer",
			Note:      "внутренний перевод",
			CreatedAt: time.Now(),
		}
		return s.transactionRepo.CreateTx(tx, tr)
	})
}

// пополнение счёта
func (s *BankService) Deposit(req models.DepositRequest) error {
	if req.Amount.LessOrEqual(decimal.Zero) {
		return errors.New("сумма должна быть >0")
	}
	return s.accountRepo.WithTx(func(tx repo.TxContext) error {
		acc, err := s.accountRepo.GetByID(req.ToAccountID)
		if err != nil {
			return err
		}
		newBal := acc.Balance.Add(req.Amount)
		if err := s.accountRepo.UpdateBalanceTx(tx, acc.ID, newBal); err != nil {
			return err
		}
		tr := &models.Transaction{
			From:      nil,
			To:        &acc.ID,
			Amount:    req.Amount,
			Type:      "deposit",
			Note:      "пополнение счёта",
			CreatedAt: time.Now(),
		}
		return s.transactionRepo.CreateTx(tx, tr)
	})
}

// оформление кредита и генерация графика
func (s *BankService) ApplyCredit(userID uuid.UUID, req models.ApplyCreditRequest) (*models.Credit, []models.PaymentSchedule, error) {
	if req.Principal.LessOrEqual(decimal.Zero) {
		return nil, nil, errors.New("сумма кредита должна быть >0")
	}
	if req.TermMonths <= 0 {
		return nil, nil, errors.New("срок должен быть >0 месяцев")
	}
	rateF, err := fetchCBRRate(time.Now())
	if err != nil || rateF <= 0 {
		rateF = 0.12 // 12% (в мечтах)) по умолчанию
	}
	annual := decimal.NewFromFloat(rateF)
	monthlyRate := annual.Div(decimal.NewFromInt(12))

	pow := monthlyRate.Add(decimal.NewFromInt(1)).Pow(decimal.NewFromInt(int64(req.TermMonths)))
	num := req.Principal.Mul(monthlyRate).Mul(pow)
	den := pow.Sub(decimal.NewFromInt(1))
	annuity := num.Div(den).Round(2)

	credit := &models.Credit{
		UserID:     userID,
		AccountID:  req.AccountID,
		Principal:  req.Principal,
		AnnualRate: annual,
		TermMonths: req.TermMonths,
		Remaining:  req.Principal,
	}

	schedules := make([]models.PaymentSchedule, 0, req.TermMonths)
	nextDue := time.Now().AddDate(0, 1, 0)
	remaining := req.Principal

	// запускаем транзакцию
	err = s.creditRepo.WithTx(func(tx repo.TxContext) error {
		if err := s.creditRepo.CreateTx(tx, credit); err != nil {
			return err
		}
		for i := 0; i < req.TermMonths; i++ {
			interest := remaining.Mul(monthlyRate).Round(2)
			principalPart := annuity.Sub(interest).Round(2)
			remaining = remaining.Sub(principalPart).Round(2)
			sched := &models.PaymentSchedule{
				CreditID:  credit.ID,
				DueDate:   nextDue,
				Amount:    annuity,
				Principal: principalPart,
				Interest:  interest,
			}
			if err := s.scheduleRepo.CreateTx(tx, sched); err != nil {
				return err
			}
			schedules = append(schedules, *sched)
			nextDue = nextDue.AddDate(0, 1, 0)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return credit, schedules, nil
}

// список кредитов пользователя
func (s *BankService) GetCredits(userID uuid.UUID) ([]models.Credit, error) {
	return s.creditRepo.GetByUserID(userID)
}

// график платежей по кредиту
func (s *BankService) GetSchedule(creditID uuid.UUID) ([]models.PaymentSchedule, error) {
	return s.scheduleRepo.GetByCreditID(creditID)
}

// списание по кредиту
func (s *BankService) ProcessScheduledPayments() error {
	today := time.Now().UTC()
	dueList, err := s.scheduleRepo.GetDueSchedules(today)
	if err != nil {
		return err
	}
	for _, sch := range dueList {
		err := s.accountRepo.WithTx(func(tx repo.TxContext) error {
			cr, err := s.creditRepo.GetByID(sch.CreditID)
			if err != nil {
				return err
			}

			acc, err := s.accountRepo.GetByID(cr.AccountID)
			if err != nil {
				return err
			}
			if acc.Balance.LessThan(sch.Amount) {
				logrus.Warnf("нехватка средств для графика %s", sch.ID)
				return nil
			}
			// списываем баланс
			newBal := acc.Balance.Sub(sch.Amount)
			if err := s.accountRepo.UpdateBalanceTx(tx, acc.ID, newBal); err != nil {
				return err
			}
			// записываем транзакцию
			tr := &models.Transaction{
				From:      &acc.ID,
				To:        nil,
				Amount:    sch.Amount,
				Type:      "credit_payment",
				Note:      fmt.Sprintf("очередной платёж по кредиту %s", sch.CreditID),
				CreatedAt: time.Now(),
			}
			if err := s.transactionRepo.CreateTx(tx, tr); err != nil {
				return err
			}
			// помечаем как оплачено
			return s.scheduleRepo.UpdatePaidTx(tx, sch.ID, true)
		})
		if err != nil {
			logrus.Errorf("ошибка шедулера: %v", err)
		}
	}
	return nil
}
