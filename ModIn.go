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
	"time"
)

type VersionConfig struct {
	MinecraftVersion string `json:"minecraft_version"`
	Loader           string `json:"loader"`
}

type CartItem struct {
	ProjectID   string
	ProjectName string
}

var (
	ROOT       string
	PACKS      string
	MC         string
	LOGFILE    string
	COUNT      int
	MODS       int
	CHOICE     string
	NAME       string
	CHOOSE     string
	TARGET     string
	DUPLICATE  string
	Cart       []CartItem
)

// Developer Console Logger (Ghi vào file log và in ra nếu cần)
func DevLog(format string, a ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	msg := fmt.Sprintf("[%s] [DEV CONSOLE] "+format+"\n", append([]interface{}{timestamp}, a...)...)
	
	// Ghi log vào file để cửa sổ console riêng đọc được real-time
	f, err := os.OpenFile(LOGFILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		defer f.Close()
		f.WriteString(msg)
	}
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	ROOT = filepath.Dir(exe)
	PACKS = filepath.Join(ROOT, "Modpacks")
	LOGFILE = filepath.Join(ROOT, "dev_console.log")
	
	// Reset file log khi khởi động mới
	os.WriteFile(LOGFILE, []byte("=== MODSIN DEVELOPER CONSOLE STARTED ===\n"), 0644)

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
		fmt.Printf("Mod Cart     : %d mods selected\n", len(Cart))
		fmt.Println()
		fmt.Println("===== Current Mods in .minecraft =====")
		files, _ := filepath.Glob(filepath.Join(MC, "*.jar"))
		for _, file := range files {
			fmt.Println("- " + filepath.Base(file))
		}
		fmt.Println()
		fmt.Println("======================================")
		fmt.Println()
		fmt.Println("1. Create Modpack")
		fmt.Println("2. List Modpacks")
		fmt.Println("3. Switch Modpack")
		fmt.Println("4. Mod Store / Search & Add to Cart (Modrinth)")
		fmt.Println("5. View Cart & Download All")
		fmt.Println("0. Exit")
		fmt.Println()
		fmt.Println("[Type 'SC1' to open separate Developer Console]")
		fmt.Println()
		fmt.Print("Select option: ")
		reader := bufio.NewReader(os.Stdin)
		CHOICE, _ = reader.ReadString('\n')
		CHOICE = strings.TrimSpace(CHOICE)

		// Lệnh mở cửa sổ Developer Console riêng khi nhập SC1 (tương đương Shift+C+1)
		if strings.EqualFold(CHOICE, "SC1") {
			OpenDevConsoleWindow()
			continue
		}

		switch CHOICE {
		case "1":
			Create()
		case "2":
			List()
		case "3":
			Switch()
		case "4":
			ModrinthSearchMenu()
		case "5":
			ViewCartAndDownload()
		case "0":
			return
		}
	}
}

// Hàm bật một cửa sổ CMD/PowerShell riêng để theo dõi log trực tiếp (Tail log file)
func OpenDevConsoleWindow() {
	// Đảm bảo file log tồn tại trước khi mở cửa sổ mới
	if _, err := os.Stat(LOGFILE); os.IsNotExist(err) {
		os.WriteFile(LOGFILE, []byte("=== MODSIN DEVELOPER CONSOLE STARTED ===\n"), 0644)
	}

	DevLog("Opening separate Developer Console window...")
	
	// Dùng lệnh PowerShell có vòng lặp kiểm tra và chờ file log sẵn sàng, tránh bị tắt ngang
	psCommand := fmt.Sprintf(`
		while (!(Test-Path '%s')) { Start-Sleep -Milliseconds 200 }
		Get-Content -Path '%s' -Wait
	`, LOGFILE, LOGFILE)

	cmd := exec.Command("cmd", "/c", "start", "cmd", "/k", "title ModsIn Developer Console & powershell -Command \""+psCommand+"\"")
	cmd.Start()
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
	fmt.Println("          ModsIn v3.4 - Dev Edition")
	fmt.Println("        Minecraft Modpack Manager")
	fmt.Println("============================================")
	fmt.Println()
}

func Create() {
	Header()
	fmt.Println("=== Create Modpack ===")
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
		fmt.Println("\nModpack already exists.")
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

	config := VersionConfig{
		MinecraftVersion: mcVer,
		Loader:           loader,
	}
	fileData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(packPath, "version.json"), fileData, 0644)

	DevLog("Created modpack folder: %s", packPath)
	fmt.Println("\nModpack created successfully!")
	Pause()
}

func List() {
	Header()
	fmt.Println("=== List Modpacks ===")
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
	fmt.Println("=== Switch Modpack ===")
	fmt.Println()
	ID := 0
	Names := make(map[int]string)
	dirs, _ := os.ReadDir(PACKS)
	for _, dir := range dirs {
		if dir.IsDir() {
			ID++
			Names[ID] = dir.Name()
			
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
	fmt.Print("Select Modpack: ")
	reader := bufio.NewReader(os.Stdin)
	CHOOSE, _ = reader.ReadString('\n')
	CHOOSE = strings.TrimSpace(CHOOSE)
	var index int
	fmt.Sscanf(CHOOSE, "%d", &index)
	TARGET = Names[index]
	if TARGET == "" {
		fmt.Println("\nInvalid selection.")
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
	fmt.Println("\nCurrent Mods in directory:")
	for _, file := range files {
		fmt.Println("- " + filepath.Base(file))
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
	DevLog("Copied mods for modpack: %s", TARGET)
	fmt.Printf("\nSwitched to %s successfully.\n", TARGET)
	Pause()
}

func Duplicate() {
	DevLog("Matching duplicate detected: %s", DUPLICATE)
	fmt.Printf("\nCurrent mods match Modpack '%s'. Switching...\n", DUPLICATE)
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
		fmt.Println("\nModpack already exists.")
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
	fmt.Println("\nModpack saved and switched successfully.")
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
	fmt.Printf("\nSwitched to %s successfully.\n", TARGET)
	Pause()
}

// === MODRINTH HUB & CART ===

func ModrinthSearchMenu() {
	Header()
	fmt.Println("=== Select Target Modpack ===")
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
	fmt.Print("Select Modpack number: ")
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
		cfg = VersionConfig{MinecraftVersion: "1.20.1", Loader: "fabric"}
	}

	ModrinthHub(targetPack, cfg)
}

func ModrinthHub(packName string, cfg VersionConfig) {
	reader := bufio.NewReader(os.Stdin)

	facets := fmt.Sprintf(`[["project_type:mod"], ["versions:%s"], ["categories:%s"]]`, cfg.MinecraftVersion, cfg.Loader)
	popularURL := fmt.Sprintf("https://api.modrinth.com/v2/search?index=downloads&limit=8&facets=%s", url.QueryEscape(facets))

	DevLog("Fetching popular mods from Modrinth (MC: %s, Loader: %s)", cfg.MinecraftVersion, cfg.Loader)

	var popularHits []struct {
		Title       string `json:"title"`
		ProjectID   string `json:"project_id"`
		Description string `json:"description"`
	}

	req, err := http.NewRequest("GET", popularURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "ModsInManager/1.0 (contact@modsin.local)")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			var result struct {
				Hits []struct {
					Title       string `json:"title"`
					ProjectID   string `json:"project_id"`
					Description string `json:"description"`
				} `json:"hits"`
			}
			json.NewDecoder(resp.Body).Decode(&result)
			popularHits = result.Hits
			DevLog("Successfully fetched %d popular mods", len(popularHits))
		} else {
			DevLog("HTTP Error fetching popular mods: %v", err)
		}
	}

	for {
		Header()
		fmt.Printf("Target Modpack: [ %s ] (MC: %s | Loader: %s)\n", packName, cfg.MinecraftVersion, cfg.Loader)
		fmt.Printf("Cart Items: %d mods\n", len(Cart))
		fmt.Println("--------------------------------------------")
		fmt.Println("🔥 POPULAR MODS (Quick Add):")
		if len(popularHits) == 0 {
			fmt.Println("(Could not load popular mods or none found)")
		} else {
			for i, hit := range popularHits {
				fmt.Printf("[%d] %s\n    -> %s\n", i+1, hit.Title, hit.Description)
			}
		}
		fmt.Println("--------------------------------------------")
		fmt.Println("[s] Search mods by name")
		fmt.Println("[0] Back to Main Menu")
		fmt.Println()
		fmt.Print("Enter option number or 's': ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "0" {
			return
		}
		if strings.EqualFold(input, "s") {
			SearchScreen(packName, cfg)
			continue
		}

		var choiceIdx int
		_, err := fmt.Sscanf(input, "%d", &choiceIdx)
		if err == nil && choiceIdx > 0 && choiceIdx <= len(popularHits) {
			selected := popularHits[choiceIdx-1]
			Cart = append(Cart, CartItem{ProjectID: selected.ProjectID, ProjectName: selected.Title})
			DevLog("Added to cart: %s (%s)", selected.Title, selected.ProjectID)
			fmt.Printf("\nAdded [%s] to cart!\n", selected.Title)
			Pause()
		}
	}
}

func SearchScreen(packName string, cfg VersionConfig) {
	reader := bufio.NewReader(os.Stdin)
	for {
		Header()
		fmt.Printf("SEARCH MODS for [ %s ] (MC: %s | Loader: %s)\n", packName, cfg.MinecraftVersion, cfg.Loader)
		fmt.Printf("Cart Items: %d mods\n", len(Cart))
		fmt.Println("--------------------------------------------")
		fmt.Print("Enter mod name to search (or type 'back' to return): ")
		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		if strings.EqualFold(query, "back") {
			return
		}

		normalizedQuery := strings.ToLower(query)
		normalizedQuery = strings.ReplaceAll(normalizedQuery, " ", "-")

		facets := fmt.Sprintf(`[["project_type:mod"], ["versions:%s"], ["categories:%s"]]`, cfg.MinecraftVersion, cfg.Loader)
		apiURL := fmt.Sprintf("https://api.modrinth.com/v2/search?query=%s&facets=%s", url.QueryEscape(normalizedQuery), url.QueryEscape(facets))

		DevLog("Searching Modrinth query: %s", normalizedQuery)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			DevLog("Error creating search request: %v", err)
			Pause()
			continue
		}
		req.Header.Set("User-Agent", "ModsInManager/1.0 (contact@modsin.local)")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			DevLog("Connection error: %v", err)
			fmt.Println("Connection error to Modrinth API.")
			Pause()
			continue
		}

		var result struct {
			Hits []struct {
				Title       string `json:"title"`
				ProjectID   string `json:"project_id"`
				Description string `json:"description"`
			} `json:"hits"`
		}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		if len(result.Hits) == 0 {
			fmt.Println("\nNo mods found matching your query.")
			Pause()
			continue
		}

		for {
			Header()
			fmt.Printf("SEARCH RESULTS FOR: '%s'\n", query)
			fmt.Printf("Cart Items: %d mods\n", len(Cart))
			fmt.Println("--------------------------------------------")
			for i, hit := range result.Hits {
				fmt.Printf("[%d] %s\n    -> %s\n", i+1, hit.Title, hit.Description)
			}
			fmt.Println("--------------------------------------------")
			fmt.Println("[0] Search another keyword")
			fmt.Println()
			fmt.Print("Select number to ADD TO CART (0 to change keyword): ")

			choiceStr, _ := reader.ReadString('\n')
			choiceStr = strings.TrimSpace(choiceStr)
			var choiceIdx int
			fmt.Sscanf(choiceStr, "%d", &choiceIdx)

			if choiceIdx == 0 {
				break
			}
			if choiceIdx > 0 && choiceIdx <= len(result.Hits) {
				selected := result.Hits[choiceIdx-1]
				Cart = append(Cart, CartItem{ProjectID: selected.ProjectID, ProjectName: selected.Title})
				DevLog("Added to cart: %s (%s)", selected.Title, selected.ProjectID)
				fmt.Printf("\nAdded [%s] to cart!\n", selected.Title)
				Pause()
			}
		}
	}
}

func ViewCartAndDownload() {
	Header()
	fmt.Println("=== MOD CART & DOWNLOADER ===")
	if len(Cart) == 0 {
		fmt.Println("Cart is empty!")
		Pause()
		return
	}

	for i, item := range Cart {
		fmt.Printf("%d. %s\n", i+1, item.ProjectName)
	}

	fmt.Println()
	fmt.Print("Select target Modpack to download all mods into: ")
	
	dirs, _ := os.ReadDir(PACKS)
	ID := 0
	Names := make(map[int]string)
	for _, dir := range dirs {
		if dir.IsDir() {
			ID++
			Names[ID] = dir.Name()
			fmt.Printf("[%d] %s\n", ID, dir.Name())
		}
	}
	
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	var index int
	fmt.Sscanf(strings.TrimSpace(input), "%d", &index)
	targetPack := Names[index]

	if targetPack == "" {
		fmt.Println("Modpack does not exist.")
		Pause()
		return
	}

	vPath := filepath.Join(PACKS, targetPack, "version.json")
	var cfg VersionConfig
	if data, err := os.ReadFile(vPath); err == nil {
		json.Unmarshal(data, &cfg)
	} else {
		cfg = VersionConfig{MinecraftVersion: "1.20.1", Loader: "fabric"}
	}

	DevLog("Starting batch download for Modpack: %s", targetPack)
	fmt.Println("\nStarting batch download process...")

	downloadQueue := make([]CartItem, len(Cart))
	copy(downloadQueue, Cart)
	processedIDs := make(map[string]bool)

	for i := 0; i < len(downloadQueue); i++ {
		item := downloadQueue[i]
		if processedIDs[item.ProjectID] {
			continue
		}
		processedIDs[item.ProjectID] = true

		DevLog("Processing download queue item: %s (%s)", item.ProjectName, item.ProjectID)
		fmt.Printf("\nDownloading: %s ...\n", item.ProjectName)
		
		depIDs := DownloadModVersionAndGetDeps(item.ProjectID, targetPack, cfg)
		
		for _, depID := range depIDs {
			if !processedIDs[depID] {
				DevLog("Discovered required dependency: %s", depID)
				downloadQueue = append(downloadQueue, CartItem{ProjectID: depID, ProjectName: "Dependency (" + depID + ")"})
			}
		}
	}

	DevLog("Batch download completed successfully.")
	fmt.Println("\nAll mods downloaded successfully!")
	Cart = nil
	Pause()
}

func DownloadModVersionAndGetDeps(projectID string, packName string, cfg VersionConfig) []string {
	versionsURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", projectID)
	DevLog("Fetching versions from: %s", versionsURL)

	req, err := http.NewRequest("GET", versionsURL, nil)
	if err != nil {
		DevLog("Error creating version request: %v", err)
		return nil
	}
	req.Header.Set("User-Agent", "ModsInManager/1.0 (contact@modsin.local)")
	
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		DevLog("Failed to fetch versions: %v", err)
		fmt.Println(" -> Connection error while fetching versions.")
		return nil
	}
	defer resp.Body.Close()

	var versions []struct {
		Name        string `json:"name"`
		VersionType string `json:"version_type"`
		Files       []struct {
			URL      string `json:"url"`
			Filename string `json:"filename"`
		} `json:"files"`
		GameVersions []string `json:"game_versions"`
		Loaders      []string `json:"loaders"`
		Dependencies []struct {
			ProjectID      string `json:"project_id"`
			DependencyType string `json:"dependency_type"`
		} `json:"dependencies"`
	}

	json.NewDecoder(resp.Body).Decode(&versions)

	var targetFileUrl, targetFileName string
	var requiredDeps []string

	for _, v := range versions {
		if v.VersionType != "release" {
			continue
		}
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
			for _, dep := range v.Dependencies {
				if dep.DependencyType == "required" && dep.ProjectID != "" {
					requiredDeps = append(requiredDeps, dep.ProjectID)
				}
			}
			break
		}
	}

	if targetFileUrl == "" {
		for _, v := range versions {
			if v.VersionType == "release" {
				continue
			}
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
				for _, dep := range v.Dependencies {
					if dep.DependencyType == "required" && dep.ProjectID != "" {
						requiredDeps = append(requiredDeps, dep.ProjectID)
					}
				}
				break
			}
		}
	}

	if targetFileUrl == "" {
		DevLog("No compatible version found for project ID: %s", projectID)
		fmt.Println(" -> No compatible version found for this MC version/loader.")
		return nil
	}

	DevLog("Downloading file: %s from URL: %s", targetFileName, targetFileUrl)

	outPath := filepath.Join(PACKS, packName, targetFileName)
	out, err := os.Create(outPath)
	if err != nil {
		DevLog("Failed to create file on disk: %v", err)
		fmt.Println(" -> Error creating file:", err)
		return nil
	}
	defer out.Close()

	fileReq, err := http.NewRequest("GET", targetFileUrl, nil)
	if err != nil {
		DevLog("Failed to create file download request: %v", err)
		return nil
	}
	fileReq.Header.Set("User-Agent", "ModsInManager/1.0 (contact@modsin.local)")

	fileResp, err := client.Do(fileReq)
	if err != nil {
		DevLog("Failed to download file from server: %v", err)
		fmt.Println(" -> Error downloading file from server:", err)
		return nil
	}
	defer fileResp.Body.Close()

	_, err = io.Copy(out, fileResp.Body)
	if err != nil {
		DevLog("Error saving file stream: %v", err)
		fmt.Println(" -> Error saving file stream:", err)
		return nil
	}

	DevLog("Successfully downloaded: %s", targetFileName)
	fmt.Printf(" -> Downloaded successfully: %s\n", targetFileName)

	return requiredDeps
}

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
