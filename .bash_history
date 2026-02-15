pkg install python
python
clear
nano kalkulator_sederhana.py
ls
run
run kalkulator_sederhana.py
# Kalkulator Sederhana
print("Kalkulator Sederhana")
print("1. Tambah")
print("2. Kurang")
pilihan = input("Masukkan pilihan (1/2): ")
angka1 = float(input("Masukkan angka pertama: "))
angka2 = float(input("Masukkan angka kedua: "))
if pilihan == '1':;     hasil = angka1 + angka2;     print(f"Hasilnya adalah {hasil}")
elif pilihan == '2':
else:
clear
# Update dulu package manager
pkg update && pkg upgrade
# Install Python
pkg install python
# Coba install dari repository lain
pkg install python-dev python-static
# Cek apakah python sudah ada
python --version
python3 --version
# Kadang sudah terinstall secara default
pkg install nmap net-tools
# Scan jaringan Anda sendiri
nmap -sn 192.168.1.0/24
# Monitor koneksi
netstat -tulpn
pkg install nmap net-tools
# Scan jaringan Anda sendiri
nmap -sn 192.168.1.0/24
# Monitor koneksi
netstat -tulpn
# Install security tools (UNTUK BELAJAR!)
pkg install nmap hydra sqlmap
# Practice di lab sendiri
nmap -A scanme.nmap.org
# Generate password kuat
apg -m 12 -n 5
# Check hash password
echo -n "password" | md5sum
# Game sederhana dengan Python
nano game.py
run game.py
# Install tools dasar dulu
pkg update && pkg upgrade
pkg install git curl wget nmap
# Install tools recon
pkg install whois dnsutils nmap
# 1. WHOIS lookup (info domain)
whois google.com
# 2. DNS lookup
nslookup google.com
dig google.com
# 3. Ping sweep
ping -c 4 google.com
# Cek IP sendiri
ifconfig
# Scan perangkat di jaringan ANDA SENDIRI
nmap -sn 192.168.1.0/24
# Scan port di device ANDA SENDIRI
nmap -p 1-1000 192.168.1.1
# Install hash tools
pkg install hashdeep
# Generate hash dari text
echo -n "password123" | md5sum
echo -n "password123" | sha256sum
# Bandingkan hash
echo -n "hello" | md5sum
nano password_checker.sh
python password_cheker.sh
pyhton password_checker.py
python password_checker.sh py
nano password_checker.sh
python bot_keuangan.py
