package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type VersionConfig struct {
	MinecraftVersion string `json:"minecraft_version"`
	Loader           string `json:"loader"`
}

var (
	ROOT      string
	PACKS     string
	MC        string
	COUNT     int
	MODS      int
	CHOICE    string
	NAME      string
	CHOOSE    string
	TARGET    string
	DUPLICATE string
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	ROOT = filepath.Dir(exe)
	PACKS = filepath.Join(ROOT, "Modpacks")
	appdata := os.Getenv("APPDATA")
	MC = filepath.Join(appdata, ".minecraft", "mods")
	os.MkdirAll(PACKS, os.ModePerm)
	os.MkdirAll(MC, os.ModePerm)
	Menu()
}

func Menu() {
	for {
		Count()
		Header()
		fmt.Printf("Modpacks     : %d\n", COUNT)
		fmt.Printf("Mods Folder  : %s\n", MC)
		fmt.Printf("Current Mods : %d\n", MODS)
		fmt.Println()
		fmt.Println("===== Current Mods =====")
		files, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
		for _, file := range files {
			fmt.Println(filepath.Base(file))
		}
		fmt.Println()
		fmt.Println("==========================")
		fmt.Println()
		fmt.Println("1. Create Modpack")
		fmt.Println("2. List Modpacks")
		fmt.Println("3. Switch Modpack")
		fmt.Println("4. Add Mods from Modrinth (Tải mod trực tiếp)")
		fmt.Println("0. Exit")
		fmt.Println()
		fmt.Print("Select: ")
		reader := bufio.NewReader(os.Stdin)
		CHOICE, _ = reader.ReadString('\n')
		CHOICE = strings.TrimSpace(CHOICE)

		switch CHOICE {
		case "1":
			Create()
		case "2":
			List()
		case "3":
			Switch()
		case "4":
			ModrinthSearchMenu()
		case "0":
			return
		}
	}
}

func Count() {
	COUNT = 0
	dirs, _ := os.ReadDir(PACKS)
	for _, dir := range dirs {
		if dir.IsDir() {
			COUNT++
		}
	}
	MODS = 0
	files, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
	for range files {
		MODS++
	}
}

func Header() {
	ClearScreen()
	fmt.Println("============================================")
	fmt.Println("            ModsIn v2.0 - Modrinth")
	fmt.Println("        Minecraft Modpack Manager")
	fmt.Println("============================================")
	fmt.Println()
}

func Create() {
	Header()
	fmt.Println("Create Modpack")
	fmt.Println()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Modpack Name: ")
	NAME, _ = reader.ReadString('\n')
	NAME = strings.TrimSpace(NAME)
	if NAME == "" {
		return
	}

	packPath := filepath.Join(PACKS, NAME)
	if _, err := os.Stat(packPath); err == nil {
		fmt.Println()
		fmt.Println("Modpack already exists.")
		Pause()
		return
	}

	fmt.Print("Minecraft Version (e.g. 1.20.1): ")
	mcVer, _ := reader.ReadString('\n')
	mcVer = strings.TrimSpace(mcVer)

	fmt.Print("Loader (fabric / forge / neoforge): ")
	loader, _ := reader.ReadString('\n')
	loader = strings.ToLower(strings.TrimSpace(loader))

	os.MkdirAll(packPath, os.ModePerm)

	// Lưu file version.json
	config := VersionConfig{
		MinecraftVersion: mcVer,
		Loader:           loader,
	}
	fileData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(packPath, "version.json"), fileData, 0644)

	fmt.Println()
	fmt.Println("Created Modpack successfully with version.json!")
	fmt.Println(packPath)
	fmt.Println()
	fmt.Println("Place your .jar files into this folder or use Option 4 to download.")
	Pause()
}

func List() {
	Header()
	fmt.Println("List Modpacks")
	fmt.Println()
	dirs, _ := os.ReadDir(PACKS)
	for _, dir := range dirs {
		if dir.IsDir() {
			vFile := filepath.Join(PACKS, dir.Name(), "version.json")
			verInfo := ""
			if data, err := os.ReadFile(vFile); err == nil {
				var cfg VersionConfig
				json.Unmarshal(data, &cfg)
				verInfo = fmt.Sprintf(" [MC: %s | Loader: %s]", cfg.MinecraftVersion, cfg.Loader)
			}
			fmt.Printf("- %s%s\n", dir.Name(), verInfo)
		}
	}
	fmt.Println()
	Pause()
}

func Switch() {
	Header()
	fmt.Println("Switch Modpack")
	fmt.Println()
	ID := 0
	Names := make(map[int]string)
	dirs, _ := os.ReadDir(PACKS)
	for _, dir := range dirs {
		if dir.IsDir() {
			ID++
			Names[ID] = dir.Name()
			
			// Đọc version hiển thị cho đẹp
			vFile := filepath.Join(PACKS, dir.Name(), "version.json")
			verInfo := ""
			if data, err := os.ReadFile(vFile); err == nil {
				var cfg VersionConfig
				json.Unmarshal(data, &cfg)
				verInfo = fmt.Sprintf(" (MC: %s, %s)", cfg.MinecraftVersion, cfg.Loader)
			}
			fmt.Printf("[%d] %s%s\n", ID, dir.Name(), verInfo)
		}
	}
	fmt.Println()
	fmt.Print("Select: ")
	reader := bufio.NewReader(os.Stdin)
	CHOOSE, _ = reader.ReadString('\n')
	CHOOSE = strings.TrimSpace(CHOOSE)
	var index int
	fmt.Sscanf(CHOOSE, "%d", &index)
	TARGET = Names[index]
	if TARGET == "" {
		fmt.Println()
		fmt.Println("Invalid selection.")
		Pause()
		return
	}
	files, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
	if len(files) == 0 {
		CopyOnly()
		return
	}
	CheckDuplicate()
	if DUPLICATE != "" {
		Duplicate()
		return
	}
	fmt.Println()
	fmt.Println("Current Mods")
	fmt.Println()
	for _, file := range files {
		fmt.Println(filepath.Base(file))
	}
	fmt.Println()
	fmt.Println("[Y] Yes (Save current mods to a new modpack)")
	fmt.Println("[N] No (Delete current mods and switch)")
	fmt.Println("[C] Cancel")
	fmt.Print("> ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToUpper(answer))
	switch answer {
	case "Y":
		Save()
	case "N":
		Delete()
	default:
		return
	}
}

func CheckDuplicate() {
	DUPLICATE = ""
	MCCOUNT := 0
	mcFiles, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
	for range mcFiles {
		MCCOUNT++
	}
	dirs, _ := os.ReadDir(PACKS)
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		packPath := filepath.Join(PACKS, dir.Name())
		PACKCOUNT := 0
		packFiles, _ := filepath.Glob(filepath.Join(packPath, "*.jar"))
		for range packFiles {
			PACKCOUNT++
		}
		if PACKCOUNT == MCCOUNT {
			MATCH := true
			for _, mcFile := range mcFiles {
				fileName := filepath.Base(mcFile)
				targetFile := filepath.Join(packPath, fileName)
				targetInfo, err := os.Stat(targetFile)
				if err != nil {
					MATCH = false
					break
				}
				mcInfo, _ := os.Stat(mcFile)
				if mcInfo.Size() != targetInfo.Size() {
					MATCH = false
					break
				}
			}
			if MATCH {
				DUPLICATE = dir.Name()
				return
			}
		}
	}
}

func CopyOnly() {
	files, _ := filepath.Glob(filepath.Join(PACKS, TARGET, "*.jar"))
	for _, file := range files {
		dst := filepath.Join(MC, filepath.Base(file))
		CopyFile(file, dst)
	}
	fmt.Println()
	fmt.Printf("Switched to %s.\n", TARGET)
	Pause()
}

func Duplicate() {
	fmt.Println()
	fmt.Printf("Current mods matched Modpack '%s'. Replacing...\n", DUPLICATE)
	fmt.Println()
	files, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
	for _, file := range files {
		os.Remove(file)
	}
	packFiles, _ := filepath.Glob(filepath.Join(PACKS, TARGET, "*.jar"))
	for _, file := range packFiles {
		dst := filepath.Join(MC, filepath.Base(file))
		CopyFile(file, dst)
	}
	Pause()
}

func Save() {
	fmt.Println()
	fmt.Print("New Modpack Name: ")
	reader := bufio.NewReader(os.Stdin)
	NEWNAME, _ := reader.ReadString('\n')
	NEWNAME = strings.TrimSpace(NEWNAME)
	if NEWNAME == "" {
		return
	}
	if _, err := os.Stat(filepath.Join(PACKS, NEWNAME)); err == nil {
		fmt.Println()
		fmt.Println("Modpack already exists.")
		Pause()
		return
	}
	os.Mkdir(filepath.Join(PACKS, NEWNAME), os.ModePerm)
	files, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
	for _, file := range files {
		dst := filepath.Join(PACKS, NEWNAME, filepath.Base(file))
		os.Rename(file, dst)
	}
	packFiles, _ := filepath.Glob(filepath.Join(PACKS, TARGET, "*.jar"))
	for _, file := range packFiles {
		dst := filepath.Join(MC, filepath.Base(file))
		CopyFile(file, dst)
	}
	fmt.Println()
	fmt.Println("Modpack saved.")
	fmt.Printf("Switched to %s.\n", TARGET)
	Pause()
}

func Delete() {
	fmt.Print("Delete all current mods? (Y/N): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToUpper(answer))
	if answer != "Y" {
		return
	}
	files, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
	for _, file := range files {
		os.Remove(file)
	}
	packFiles, _ := filepath.Glob(filepath.Join(PACKS, TARGET, "*.jar"))
	for _, file := range packFiles {
		dst := filepath.Join(MC, filepath.Base(file))
		CopyFile(file, dst)
	}
	fmt.Println()
	fmt.Printf("Switched to %s.\n", TARGET)
	Pause()
}

// === TÍNH NĂNG MỚI: TẢI MOD TỪ MODRINTH DỰA TRÊN version.json ===

func ModrinthSearchMenu() {
	Header()
	fmt.Println("Select Modpack to Add Mods")
	fmt.Println()

	ID := 0
	Names := make(map[int]string)
	dirs, _ := os.ReadDir(PACKS)
	for _, dir := range dirs {
		if dir.IsDir() {
			ID++
			Names[ID] = dir.Name()
			fmt.Printf("[%d] %s\n", ID, dir.Name())
		}
	}

	if ID == 0 {
		fmt.Println("No modpacks found. Create one first!")
		Pause()
		return
	}

	fmt.Println()
	fmt.Print("Select Modpack ID: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	var index int
	fmt.Sscanf(strings.TrimSpace(input), "%d", &index)

	targetPack := Names[index]
	if targetPack == "" {
		fmt.Println("Invalid selection.")
		Pause()
		return
	}

	vPath := filepath.Join(PACKS, targetPack, "version.json")
	var cfg VersionConfig
	if data, err := os.ReadFile(vPath); err == nil {
		json.Unmarshal(data, &cfg)
	} else {
		fmt.Println("Warning: version.json not found. Defaulting to 1.20.1 / fabric...")
		cfg = VersionConfig{MinecraftVersion: "1.20.1", Loader: "fabric"}
	}

	SearchModrinth(targetPack, cfg)
}

func SearchModrinth(packName string, cfg VersionConfig) {
	reader := bufio.NewReader(os.Stdin)
	for {
		Header()
		fmt.Printf("Modpack: %s | MC: %s | Loader: %s\n", packName, cfg.MinecraftVersion, cfg.Loader)
		fmt.Println("--------------------------------------------")
		fmt.Print("Search Mod (or type 'exit' to back): ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)
		if query == "" || query == "exit" {
			return
		}

		// Chuẩn hóa chuỗi: tự động giảm tất cả chữ xuống dạng ko viết chữ hoa và thêm dấu gạch
		normalizedQuery := strings.ToLower(query)
		normalizedQuery = strings.ReplaceAll(normalizedQuery, " ", "-")

		facets := fmt.Sprintf(`[["versions:%s"], ["categories:%s"]]`, cfg.MinecraftVersion, cfg.Loader)
		apiURL := fmt.Sprintf("https://api.modrinth.com/v2/search?query=%s&facets=%s", url.QueryEscape(normalizedQuery), url.QueryEscape(facets))

		resp, err := http.Get(apiURL)
		if err != nil {
			fmt.Println("Error connecting to Modrinth API:", err)
			Pause()
			continue
		}
		defer resp.Body.Close()

		var result struct {
			Hits []struct {
				Title       string   `json:"title"`
				ProjectID   string   `json:"project_id"`
				Description string   `json:"description"`
				Versions    []string `json:"versions"`
			} `json:"hits"`
		}

		json.NewDecoder(resp.Body).Decode(&result)

		if len(result.Hits) == 0 {
			fmt.Println("\nNo mods found matching your version/loader criteria.")
			Pause()
			continue
		}

		fmt.Println("\nFound Mods:")
		for i, hit := range result.Hits {
			fmt.Printf("[%d] %s\n    -> %s\n", i+1, hit.Title, hit.Description)
		}

		fmt.Println()
		fmt.Print("Enter number to download (0 to search again): ")
		choiceStr, _ := reader.ReadString('\n')
		var choiceIdx int
		fmt.Sscanf(strings.TrimSpace(choiceStr), "%d", &choiceIdx)

		if choiceIdx <= 0 || choiceIdx > len(result.Hits) {
			continue
		}

		selectedMod := result.Hits[choiceIdx-1]
		DownloadModrinthVersion(selectedMod.ProjectID, packName, cfg)
	}
}

func DownloadModrinthVersion(projectID string, packName string, cfg VersionConfig) {
	versionsURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", projectID)
	resp, err := http.Get(versionsURL)
	if err != nil {
		fmt.Println("Error fetching mod versions:", err)
		Pause()
		return
	}
	defer resp.Body.Close()

	var versions []struct {
		Name  string `json:"name"`
		Files []struct {
			URL      string `json:"url"`
			Filename string `json:"filename"`
		} `json:"files"`
		GameVersions []string `json:"game_versions"`
		Loaders      []string `json:"loaders"`
	}

	json.NewDecoder(resp.Body).Decode(&versions)

	var targetFileUrl, targetFileName string
	for _, v := range versions {
		matchMC := false
		for _, gv := range v.GameVersions {
			if gv == cfg.MinecraftVersion {
				matchMC = true
				break
			}
		}
		matchLoader := false
		for _, l := range v.Loaders {
			if strings.EqualFold(l, cfg.Loader) {
				matchLoader = true
				break
			}
		}

		if matchMC && matchLoader && len(v.Files) > 0 {
			targetFileUrl = v.Files[0].URL
			targetFileName = v.Files[0].Filename
			break
		}
	}

	if targetFileUrl == "" {
		fmt.Println("\nCould not find a compatible version for your MC/Loader configuration.")
		Pause()
		return
	}

	fmt.Printf("\nDownloading %s...\n", targetFileName)
	outPath := filepath.Join(PACKS, packName, targetFileName)

	out, err := os.Create(outPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		Pause()
		return
	}
	defer out.Close()

	fileResp, err := http.Get(targetFileUrl)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		Pause()
		return
	}
	defer fileResp.Body.Close()

	_, err = io.Copy(out, fileResp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		Pause()
		return
	}

	fmt.Println("Download completed successfully!")
	Pause()
}

// === CÁC HÀM TIỆN ÍCH CŨ ===

func Pause() {
	fmt.Println()
	fmt.Print("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func CopyFile(src, dst string) {
	source, err := os.Open(src)
	if err != nil {
		return
	}
	defer source.Close()
	target, err := os.Create(dst)
	if err != nil {
		return
	}
	defer target.Close()
	io.Copy(target, source)
}

func ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
