#!/bin/bash
echo "Password Strength Checker"
read -s -p "Masukkan password: " password
echo

# Check length
if [ ${#password} -lt 8 ]; then
    echo "❌ Terlalu pendek (min 8 karakter)"
else
    echo "✅ Panjang cukup"
fi

# Check complexity
if [[ $password =~ [A-Z] ]] && [[ $password =~ [a-z] ]] && [[ $password =~ [0-9] ]]; then
    echo "✅ Kompleksitas baik"
else
    echo "❌ Kurang kompleks"
fi
