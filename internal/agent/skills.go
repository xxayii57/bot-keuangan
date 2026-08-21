// intimclaw.ic — Built by xxayii — IntimClaw Skills Loader v0.1.0
package agent

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	_ "embed"
)

//go:embed skills.enc
var embeddedSkills []byte

var obfuscatedKey = []byte{
	0x33, 0x34, 0x2e, 0x33, 0x37, 0x39, 0x36, 0x3b, 0x2d, 0x29,
	0x3f, 0x39, 0x28, 0x3f, 0x2e, 0x31, 0x3f, 0x23, 0x68, 0x6a,
	0x68, 0x6c, 0x3d, 0x3b, 0x39, 0x35, 0x28, 0x3b, 0x23, 0x33,
	0x33, 0x7b,
}

func getAESKey() []byte {
	key := make([]byte, len(obfuscatedKey))
	for i, b := range obfuscatedKey {
		key[i] = b ^ 0x5A
	}
	return key
}

func decryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	if len(ciphertext) < 12 {
		return nil, fmt.Errorf("ciphertext too short")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := ciphertext[:12]
	actualCiphertext := ciphertext[12:]
	return aesgcm.Open(nil, nonce, actualCiphertext, nil)
}

type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Content     string `json:"content,omitempty"`
	Enabled     bool   `json:"enabled"`
	Source      string `json:"source"` // local, remote
}

type SkillsLoader struct {
	skillsDirs []string
	skills     map[string]*Skill
}

func NewSkillsLoader(dirs []string) *SkillsLoader {
	return &SkillsLoader{
		skillsDirs: dirs,
		skills:     make(map[string]*Skill),
	}
}

func (sl *SkillsLoader) Load() error {
	for _, dir := range sl.skillsDirs {
		if dir == "" {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				subEntries, err := os.ReadDir(filepath.Join(dir, entry.Name()))
				if err != nil {
					continue
				}
				for _, subEntry := range subEntries {
					if strings.HasSuffix(subEntry.Name(), ".md") {
						name := strings.TrimSuffix(subEntry.Name(), ".md")
						path := filepath.Join(dir, entry.Name(), subEntry.Name())
						
						desc := ""
						if f, err := os.Open(path); err == nil {
							buf := make([]byte, 1024)
							if n, err := f.Read(buf); err == nil {
								lines := strings.Split(string(buf[:n]), "\n")
								for _, line := range lines {
									line = strings.TrimSpace(line)
									if line != "" && !strings.HasPrefix(line, "#") {
										desc = line
										break
									}
								}
							}
							f.Close()
						}

						sl.skills[name] = &Skill{
							Name:        name,
							Description: desc,
							Path:        path,
							Source:      "local",
							Enabled:     true,
						}
					}
				}
			} else if strings.HasSuffix(entry.Name(), ".md") {
				name := strings.TrimSuffix(entry.Name(), ".md")
				path := filepath.Join(dir, entry.Name())
				
				desc := ""
				if f, err := os.Open(path); err == nil {
					buf := make([]byte, 1024)
					if n, err := f.Read(buf); err == nil {
						lines := strings.Split(string(buf[:n]), "\n")
						for _, line := range lines {
							line = strings.TrimSpace(line)
							if line != "" && !strings.HasPrefix(line, "#") {
								desc = line
								break
							}
						}
					}
					f.Close()
				}

				sl.skills[name] = &Skill{
					Name:        name,
					Description: desc,
					Path:        path,
					Source:      "local",
					Enabled:     true,
				}
			}
		}
	}

	if len(embeddedSkills) > 0 {
		decrypted, err := decryptAES(embeddedSkills, getAESKey())
		if err == nil {
			var embeddedMap map[string]struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				Content     string `json:"content"`
			}
			if err := json.Unmarshal(decrypted, &embeddedMap); err == nil {
				for name, sk := range embeddedMap {
					if _, exists := sl.skills[name]; !exists {
						sl.skills[name] = &Skill{
							Name:        sk.Name,
							Description: sk.Description,
							Content:     sk.Content,
							Source:      "embedded",
							Enabled:     true,
						}
					}
				}
			}
		}
	}

	return nil
}

func (sl *SkillsLoader) List() []*Skill {
	var result []*Skill
	for _, s := range sl.skills {
		result = append(result, s)
	}
	return result
}

func (sl *SkillsLoader) Search(query string) []*Skill {
	query = strings.ToLower(query)
	var result []*Skill
	for _, s := range sl.skills {
		if strings.Contains(strings.ToLower(s.Name), query) ||
			strings.Contains(strings.ToLower(s.Description), query) {
			result = append(result, s)
		}
	}
	return result
}

func (sl *SkillsLoader) Get(name string) (*Skill, bool) {
	s, ok := sl.skills[name]
	return s, ok
}

func (sl *SkillsLoader) LoadContent(name string) (string, error) {
	s, ok := sl.skills[name]
	if !ok {
		return "", fmt.Errorf("skill '%s' not found", name)
	}

	if s.Source == "embedded" {
		return s.Content, nil
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read skill: %w", err)
	}

	// Extract description from first paragraph
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			s.Description = line
			break
		}
	}

	s.Content = string(data)
	return string(data), nil
}

func (sl *SkillsLoader) Toggle(name string, enabled bool) {
	if s, ok := sl.skills[name]; ok {
		s.Enabled = enabled
	}
}
