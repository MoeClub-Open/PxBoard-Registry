package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	src := flag.String("plugins", "../PxBoard-Plugins", "官方插件仓库")
	platform := flag.String("platform", runtime.GOOS+"-"+runtime.GOARCH, "发布平台")
	keyfile := flag.String("key-file", "", "可选 Ed25519 私钥文件")
	generate := flag.String("generate-key", "", "生成私钥到指定文件，输出公钥")
	artifactBase := flag.String("artifact-base", "", "安装包的公开 HTTPS 目录，例如 GitHub Release 下载地址")
	flag.Parse()
	if *artifactBase != "" {
		parsed, e := url.Parse(*artifactBase)
		if e != nil || parsed.Scheme != "https" || parsed.Host == "" {
			log.Fatal("artifact-base 必须是有效的 HTTPS 地址")
		}
	}
	if *generate != "" {
		pub, key, e := ed25519.GenerateKey(rand.Reader)
		if e != nil {
			log.Fatal(e)
		}
		if _, e = os.Stat(*generate); e == nil {
			log.Fatal("私钥文件已存在")
		}
		if e = os.WriteFile(*generate, []byte(base64.StdEncoding.EncodeToString(key)), 0600); e != nil {
			log.Fatal(e)
		}
		fmt.Println("Public key:", base64.StdEncoding.EncodeToString(pub))
		return
	}
	parts := strings.SplitN(*platform, "-", 2)
	if len(parts) != 2 {
		log.Fatal("平台格式：linux-amd64")
	}
	var key ed25519.PrivateKey
	if *keyfile != "" {
		b, e := os.ReadFile(*keyfile)
		if e != nil {
			log.Fatal(e)
		}
		key, e = base64.StdEncoding.DecodeString(strings.TrimSpace(string(b)))
		if e != nil || len(key) != ed25519.PrivateKeySize {
			log.Fatal("私钥格式无效")
		}
	}
	entries := []map[string]any{}
	if b, e := os.ReadFile("index.json"); e == nil {
		var index struct {
			Plugins []map[string]any `json:"plugins"`
		}
		if e = json.Unmarshal(b, &index); e != nil {
			log.Fatal(e)
		}
		entries = index.Plugins
	}
	byID := map[string]map[string]any{}
	for _, entry := range entries {
		byID[entry["id"].(string)] = entry
	}
	files, e := filepath.Glob(filepath.Join(*src, "plugins", "*", "manifest.json"))
	if e != nil {
		log.Fatal(e)
	}
	if len(files) == 0 {
		log.Fatal("未找到插件 manifest.json")
	}
	for _, file := range files {
		b, e := os.ReadFile(file)
		if e != nil {
			log.Fatal(e)
		}
		var entry map[string]any
		if e = json.Unmarshal(b, &entry); e != nil {
			log.Fatal(e)
		}
		id := entry["id"].(string)
		version := entry["version"].(string)
		name := id + "-" + version + "-" + *platform
		if parts[0] == "windows" {
			name += ".exe"
		}
		dest := filepath.Join("packages", name)
		if e = os.MkdirAll("packages", 0755); e != nil {
			log.Fatal(e)
		}
		absolute, e := filepath.Abs(dest)
		if e != nil {
			log.Fatal(e)
		}
		cmd := exec.Command("go", "build", "-trimpath", "-o", absolute, "./cmd/"+id)
		cmd.Dir = *src
		cmd.Env = append(os.Environ(), "GOOS="+parts[0], "GOARCH="+parts[1], "CGO_ENABLED=0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		fmt.Println("Building", id, *platform)
		if e = cmd.Run(); e != nil {
			log.Fatal(e)
		}
		artifacts := map[string]any{}
		signatures := map[string]any{}
		if old := byID[id]; old != nil && old["version"] == version {
			if v, ok := old["artifacts"].(map[string]any); ok {
				artifacts = v
			}
			if v, ok := old["signatures"].(map[string]any); ok {
				signatures = v
			}
		}
		artifact := "packages/" + name
		if *artifactBase != "" {
			artifact = strings.TrimRight(*artifactBase, "/") + "/" + url.PathEscape(name)
		}
		artifacts[*platform] = artifact
		entry["artifacts"] = artifacts
		if len(key) > 0 {
			binary, e := os.ReadFile(dest)
			if e != nil {
				log.Fatal(e)
			}
			signatures[*platform] = base64.StdEncoding.EncodeToString(ed25519.Sign(key, binary))
		}
		entry["signatures"] = signatures
		byID[id] = entry
	}
	entries = []map[string]any{}
	for _, file := range files {
		id := filepath.Base(filepath.Dir(file))
		entries = append(entries, byID[id])
	}
	var bundles []map[string]any
	if b, e := os.ReadFile("bundles.json"); e == nil {
		if e = json.Unmarshal(b, &bundles); e != nil {
			log.Fatal(e)
		}
	}
	out, e := json.MarshalIndent(map[string]any{"schemaVersion": "1.0", "name": "PxBoard 官方插件源", "plugins": entries, "bundles": bundles}, "", "  ")
	if e != nil {
		log.Fatal(e)
	}
	if e = os.WriteFile("index.json", out, 0644); e != nil {
		log.Fatal(e)
	}
	fmt.Println("Published", len(entries), "plugins")
}
