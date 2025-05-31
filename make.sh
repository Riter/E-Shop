#!/bin/bash

# Поиск всех директорий, содержащих go.mod
find . -type f -name "go.mod" | while read modfile; do
    dir=$(dirname "$modfile")
    echo ">>> Обработка директории: $dir"

    pushd "$dir" > /dev/null

    echo "→ Выполняем go mod tidy"
    go mod tidy

    echo "→ Выполняем go mod vendor"
    go mod vendor

    popd > /dev/null
    echo ""
done

echo "Все go-модули обновлены"
