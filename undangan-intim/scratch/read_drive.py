import urllib.request
import re
import codecs

url = "https://drive.google.com/drive/folders/1HhLO7BPtKy4w4mN5hKGDDPvAKsw8T__Q"
req = urllib.request.Request(
    url, 
    headers={'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'}
)
try:
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
        
        # Find _DRIVE_ivd pattern
        match = re.search(r"window\['_DRIVE_ivd'\]\s*=\s*'([^']+)'", html)
        if match:
            encoded_str = match.group(1)
            # Unescape hex values
            decoded_str = codecs.escape_decode(bytes(encoded_str, "utf-8"))[0].decode("utf-8")
            
            print("Successfully decoded _DRIVE_ivd data!")
            # Find all files in the format ["ID", ["PARENT_ID"], "NAME", "MIME_TYPE"]
            files = re.findall(r'\["([a-zA-Z0-9_-]{28,35})",\["[^"]+"\],"([^"]+)"', decoded_str)
            if not files:
                # Try generic regex to find drive file ID followed by name
                files = re.findall(r'["\']([a-zA-Z0-9_-]{33})["\'],\[.*?["\']([^"\']+\.mp3)["\']', decoded_str)
            
            print(f"Parsed {len(files)} files:")
            for fid, name in files:
                print(f"- Name: {name}")
                print(f"  ID: {fid}")
                print(f"  Direct Link: https://docs.google.com/uc?export=download&id={fid}")
        else:
            print("Could not find window['_DRIVE_ivd'] block in html.")
            # Fallback regex
            files = re.findall(r'["\'](1[a-zA-Z0-9_-]{32})["\'],\[.*?["\']([^"\']+\.mp3)["\']', html)
            print(f"Fallback matched {len(files)} files:")
            for fid, name in files:
                print(f"- {name} ({fid})")
except Exception as e:
    print("Error:", e)
