package terminal

import (
	"fmt"
	"math/rand"
	"time"
)

// Repository - интерфейс (контракт), по которому мы работаем с данными
type Repository interface {
	// EnrichATM - наполняет банкомат Forte детальной внутренней статистикой
	EnrichATM(atm *ATM)

	// EnrichCompetitor - наполняет банкомат конкурента оценочной аналитикой
	EnrichCompetitor(atm *ATM)

	// GenerateRandomCompetitors - создает фейковые точки, если OpenStreetMap недоступен
	GenerateRandomCompetitors(count int) []ATM
}

// MockRepository - имитация базы данных
type MockRepository struct{}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

// --- 1. ЛОГИКА ДЛЯ FORTE (Детальная) ---
func (r *MockRepository) EnrichATM(atm *ATM) {
	// Присваиваем признаки Forte
	atm.IsForte = true
	atm.Bank = "Forte Bank"
	// Для примера ставим район (в реальности это делается через Geo-полигоны)
	atm.District = "Город"

	// Генерируем финансовые показатели
	atm.AvgCashBalanceKZT = float64(5000000 + rand.Intn(20000000)) // 5 - 25 млн
	atm.WithdrawalFreqPerDay = 50 + rand.Intn(400)                 // 50 - 450 операций
	atm.DowntimePct = rand.Float64() * 0.15                        // 0 - 15% простоя

	// Генерируем кассеты и считаем точный баланс
	var total float64
	atm.Cassettes, total = r.genCassettes()
	atm.TotalCashKZT = total

	// Генерируем жалобы
	atm.Complaints = r.genComplaints()

	// Рассчитываем эффективность на основе сгенерированных данных
	r.calcEfficiency(atm)
}

// --- 2. ЛОГИКА ДЛЯ КОНКУРЕНТОВ (Оценочная) ---
func (r *MockRepository) EnrichCompetitor(atm *ATM) {
	atm.IsForte = false

	// Генерируем ТОЛЬКО оценочные потоки (Estimated Flows)

	// Оценка Снятия: от 2 млн до 15 млн в день
	atm.EstWithdrawalKZT = float64(2000000 + rand.Intn(13000000))

	// Оценка Внесения: от 500к до 8 млн в день
	atm.EstDepositKZT = float64(500000 + rand.Intn(7500000))

	// Остальные поля (Status, Cassettes, Complaints) остаются пустыми,
	// так как у нас нет доступа к внутренней кухне конкурентов.
}

// --- 3. FALLBACK ГЕНЕРАТОР (Если нет интернета/OSM) ---
func (r *MockRepository) GenerateRandomCompetitors(count int) []ATM {
	var atms []ATM
	banks := []string{"Kaspi", "Halyk", "Jusan", "BCC", "Eurasian"}

	// Границы Астаны для рандома
	minLat, maxLat := 51.08, 51.20
	minLng, maxLng := 71.38, 71.52

	for i := 0; i < count; i++ {
		bank := banks[rand.Intn(len(banks))]
		atms = append(atms, ATM{
			ID:      9000 + i,
			Name:    fmt.Sprintf("%s ATM #%d", bank, i),
			Lat:     minLat + rand.Float64()*(maxLat-minLat),
			Lng:     minLng + rand.Float64()*(maxLng-minLng),
			IsForte: false,
			Bank:    bank,
			// Сразу заполняем оценочными данными
			EstWithdrawalKZT: float64(2000000 + rand.Intn(10000000)),
			EstDepositKZT:    float64(1000000 + rand.Intn(5000000)),
		})
	}
	return atms
}

// --- ВСПОМОГАТЕЛЬНЫЕ ПРИВАТНЫЕ МЕТОДЫ ---

// Генерация кассет (Cash-In / Cash-Out)
func (r *MockRepository) genCassettes() ([]Cassette, float64) {
	// Кассета Выдачи (Out)
	capOut := 20000000.0
	amtOut := float64(rand.Intn(int(capOut)))
	stOut := "OK"
	if amtOut < 2000000 {
		stOut = "Low (Мало денег)"
	}
	if amtOut == 0 {
		stOut = "Empty (Пусто)"
	}

	// Кассета Приема (In)
	capIn := 10000000.0
	amtIn := float64(rand.Intn(int(capIn)))
	stIn := "OK"
	if amtIn > 9000000 {
		stIn = "Full (Переполнен)"
	}

	list := []Cassette{
		{Type: "Cash-Out", Currency: "KZT", Amount: amtOut, Capacity: capOut, Status: stOut},
		{Type: "Cash-In", Currency: "KZT", Amount: amtIn, Capacity: capIn, Status: stIn},
	}
	return list, amtOut + amtIn
}

// Генерация случайных жалоб
func (r *MockRepository) genComplaints() []Complaint {
	// У 70% терминалов жалоб нет
	if rand.Float32() > 0.3 {
		return nil
	}

	templates := []struct{ cat, txt string }{
		{"Техническая", "Зажевал карту"},
		{"Техническая", "Не выдал чек"},
		{"Техническая", "Экран не реагирует на нажатия"},
		{"Чистота", "Грязная клавиатура"},
		{"Чистота", "Мусор возле урны"},
		{"Обслуживание", "Долго обрабатывает запрос"},
		{"Обслуживание", "Нет наличных (купюры по 2000)"},
	}

	// 1 или 2 жалобы
	count := 1 + rand.Intn(2)
	var res []Complaint
	for i := 0; i < count; i++ {
		t := templates[rand.Intn(len(templates))]
		res = append(res, Complaint{
			ID:       rand.Intn(99999),
			Category: t.cat,
			Text:     t.txt,
			Status:   "Open",
			Date:     time.Now().AddDate(0, 0, -rand.Intn(10)).Format("2006-01-02"),
		})
	}
	return res
}

// Расчет эффективности (Effective / Ineffective / Normal)
func (r *MockRepository) calcEfficiency(atm *ATM) {
	// Проверяем критические статусы кассет
	cashOutStatus := atm.Cassettes[0].Status
	cashInStatus := atm.Cassettes[1].Status

	if cashOutStatus == "Empty (Пусто)" || cashInStatus == "Full (Переполнен)" || atm.DowntimePct > 0.10 {
		// Если нет денег, переполнен или часто ломается
		atm.EfficiencyStatus = "Ineffective"
	} else if atm.WithdrawalFreqPerDay > 300 && atm.DowntimePct < 0.03 {
		// Если много транзакций и редко ломается
		atm.EfficiencyStatus = "Effective"
	} else {
		// Обычный режим
		atm.EfficiencyStatus = "Normal"
	}
}
