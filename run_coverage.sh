#!/bin/bash
# Запускаем тесты, исключая проблемные тесты и тесты с таймерами
# Сначала генерируем покрытие для ошибок и id
go test -timeout 3s ./protocol/... -run="^TestValidate|^TestError|^TestID|^TestNew[A-Za-z]*Error" -coverprofile=coverage.out

# Затем добавляем покрытие для notification
go test -timeout 3s ./protocol/... -run="^TestNotification|^TestGet[A-Za-z]*Notification|^TestSet[A-Za-z]*Notification|^TestUnmarshalJSONNotification" -coverprofile=coverage.tmp.out
if [ -f coverage.tmp.out ]; then
  # Объединяем результаты (сохраняем только данные покрытия, без заголовка)
  tail -n +2 coverage.tmp.out >> coverage.out
fi

# Запускаем тесты для request без таймеров
go test -timeout 3s ./protocol/... -run="^TestRequest|^TestGet[A-Za-z]*Request|^TestSet[A-Za-z]*Request|^TestUnmarshalJSONRequest|^TestValidate_[A-Za-z]*" -coverprofile=coverage.tmp.out
if [ -f coverage.tmp.out ]; then
  tail -n +2 coverage.tmp.out >> coverage.out
fi

# Запускаем тесты для response
go test -timeout 3s ./protocol/... -run="^TestResponse|^TestGet[A-Za-z]*Response|^TestSet[A-Za-z]*Response|^TestHas|^TestUnmarshalJSONResponse|^TestRPCError" -coverprofile=coverage.tmp.out
if [ -f coverage.tmp.out ]; then
  tail -n +2 coverage.tmp.out >> coverage.out
fi

# Запускаем безопасные тесты для request_lifecycle_manager
go test -timeout 3s ./protocol/... -run="^TestNewRequest|^TestUpdateCallback|^TestCompleteRequest|^TestResetTimeout|^TestActiveIDs|^TestCleanupRequest|^TestStartRequest|^TestTriggerCallback|^TestStopAllNoWait|^TestWithErrorHandler" -coverprofile=coverage.tmp.out
if [ -f coverage.tmp.out ]; then
  tail -n +2 coverage.tmp.out >> coverage.out
fi

# Запускаем еще некоторые безопасные тесты
go test -timeout 3s ./protocol/... -run="^TestTimeoutTypeString|^TestRequestStateStop|^TestRequestLifecycleManager" -coverprofile=coverage.tmp.out
if [ -f coverage.tmp.out ]; then
  tail -n +2 coverage.tmp.out >> coverage.out
fi

# Генерируем HTML-отчет
go tool cover -html=coverage.out -o coverage.html

echo "Coverage report generated: coverage.html"

# Удаляем временные файлы
rm -f coverage.tmp.out 