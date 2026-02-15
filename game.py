import random

number = random.randint(1, 100)
guess = int(input("Tebak angka 1-100: "))

if guess == number:
    print("Benar!")
else:
    print("Salah! Angkanya:", number)
