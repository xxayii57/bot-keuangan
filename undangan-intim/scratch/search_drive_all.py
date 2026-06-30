import urllib.request
import re
import ssl

ssl._create_default_https_context = ssl._create_unverified_context

url = "https://drive.google.com/drive/folders/1HhLO7BPtKy4w4mN5hKGDDPvAKsw8T__Q"
req = urllib.request.Request(
    url,
    headers={'User-Agent': 'Mozilla/5.0'}
)

try:
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
    
    # Search for all strings matching .mp4, .mov, .jpg, .png, etc. in raw html
    raw_matches = re.findall(r'[^"\\\'\[\]\s,\(\)]+\.(?:mp4|mov|png|jpg|jpeg|avi|mkv)', html, re.IGNORECASE)
    print("Found files:", list(set(raw_matches)))
except Exception as e:
    print("Error:", e)
