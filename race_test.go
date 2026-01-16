package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

// Структуры для парсинга ответов в тесте
type LoginResponse struct {
	Token string `json:"token"`
}

func TestRaceCondition(t *testing.T) {
	// 1. НАСТРОЙКИ
	baseURL := "http://localhost:8080/api"
	roomID := 1 // Убедись, что комната с ID=1 существует в базе!

	// Время бронирования (одинаковое для всех)
	startTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(25 * time.Hour).Format(time.RFC3339)

	// 2. ПОЛУЧАЕМ ТОКЕН (Автоматически)
	token := getToken(t, baseURL)
	fmt.Println("[TEST] Токен получен. Начинаем атаку клонов...")

	// 3. ЗАПУСКАЕМ ГОНКУ
	// Эмулируем 50 одновременных запросов
	concurrentRequests := 50
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	successCount := 0
	conflictCount := 0
	errorCount := 0

	// Мьютекс нужен только чтобы красиво считать счетчики в тесте
	var mu sync.Mutex

	payload := map[string]interface{}{
		"roomId":    roomID,
		"title":     "RACE CONDITION TEST",
		"startTime": startTime,
		"endTime":   endTime,
	}
	payloadBytes, _ := json.Marshal(payload)

	for i := 0; i < concurrentRequests; i++ {
		go func(id int) {
			defer wg.Done()

			req, _ := http.NewRequest("POST", baseURL+"/bookings", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				fmt.Printf("Request %d failed: %v\n", id, err)
				errorCount++
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == 201 {
				successCount++
				fmt.Printf("✅ Горутина %d: УСПЕХ (201 Created)\n", id)
			} else if resp.StatusCode == 409 {
				conflictCount++
				// Раскомментируй, если хочешь видеть отказы
				fmt.Printf("🛡️ Горутина %d: ОТКАЗ (409 Conflict)\n", id)
			} else {
				errorCount++
				fmt.Printf("⚠️ Горутина %d: Странный код %d\n", id, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()

	// 4. ИТОГИ
	fmt.Println("------------------------------------------------")
	fmt.Printf("Всего попыток: %d\n", concurrentRequests)
	fmt.Printf("Успешных броней: %d\n", successCount)
	fmt.Printf("Отбитых атак (409): %d\n", conflictCount)
	fmt.Println("------------------------------------------------")

	if successCount == 1 && conflictCount == concurrentRequests-1 {
		t.Log("🏆 ТЕСТ ПРОЙДЕН! Транзакции работают идеально.")
	} else {
		t.Errorf("❌ ОШИБКА! Успешных броней: %d (должна быть 1). База данных не выдержала.", successCount)
	}
}

// Хелпер для логина
func getToken(t *testing.T, baseURL string) string {
	// Используем юзера, которого ты уже создал
	loginPayload := map[string]string{
		"email":    "test@user.com", // <-- Проверь, что этот юзер есть в базе
		"password": "mypassword123", // <-- И пароль верный
	}
	body, _ := json.Marshal(loginPayload)

	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Не удалось залогиниться для теста: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Ошибка логина: статус %d. Проверь test@user.com в базе", resp.StatusCode)
	}

	var res LoginResponse
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Token
}
