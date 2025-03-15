#!/bin/bash

# Компиляция C++ кода с оптимизацией
g++ script.cpp -I/usr/local/include/eigen3 -o spectral_embed -O2

# Запуск исполняемого файла с аргументами
./spectral_embed "$1" 1 1 3 "$2"

# Запуск Python-скрипта
venv/bin/python draw.py "$2"
