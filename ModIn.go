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

type CartItem struct {
	ProjectID   string
	ProjectName string
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
	Cart      []CartItem
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
		fmt.Printf("Mod Cart     : %d mods đang chọn\n", len(Cart))
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
		fmt.Println("4. Kho Mod / Tìm kiếm & Thêm vào Giỏ (Modrinth)")
		fmt.Println("5. Xem giỏ hàng & Tải xuống tất cả")
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
		case "5":
			ViewCartAndDownload()
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
	fmt.Println("            ModsIn v3.2 - Cart System")
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

	config := VersionConfig{
		MinecraftVersion: mcVer,
		Loader:           loader,
	}
	fileData, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(filepath.Join(packPath, "version.json"), fileData, 0644)

	fmt.Println()
	fmt.Println("Created Modpack successfully with version.json!")
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

// === TÍNH NĂNG KHO MOD, PHỔ BIẾN & GIỎ HÀNG ===

func ModrinthSearchMenu() {
	Header()
	fmt.Println("Chọn Modpack đích để xem kho mod:")
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
		fmt.Println("Chưa có modpack nào. Hãy tạo trước!")
		Pause()
		return
	}

	fmt.Println()
	fmt.Print("Chọn số thứ tự Modpack: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	var index int
	fmt.Sscanf(strings.TrimSpace(input), "%d", &index)

	targetPack := Names[index]
	if targetPack == "" {
		fmt.Println("Lựa chọn không hợp lệ.")
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

// Màn hình chính Modrinth (Hiển thị mod phổ biến tải nhanh)
// Màn hình chính Modrinth (Hiển thị mod phổ biến tải nhanh)
func ModrinthHub(packName string, cfg VersionConfig) {
	reader := bufio.NewReader(os.Stdin)

	facets := fmt.Sprintf(`[["project_type:mod"], ["versions:%s"], ["categories:%s"]]`, cfg.MinecraftVersion, cfg.Loader)
	popularURL := fmt.Sprintf("https://api.modrinth.com/v2/search?index=downloads&limit=8&facets=%s", url.QueryEscape(facets))

	var popularHits []struct {
		Title       string `json:"title"`
		ProjectID   string `json:"project_id"`
		Description string `json:"description"`
	}

	// Tạo request và gán User-Agent theo chuẩn Modrinth yêu cầu
	req, err := http.NewRequest("GET", popularURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "VerityApp/ModsIn/1.0 (contact@verity.gg)")
		client := &http.Client{}
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
		}
	}

	for {
		Header()
		fmt.Printf("Đang chọn mod cho Modpack: [ %s ] (MC: %s | Loader: %s)\n", packName, cfg.MinecraftVersion, cfg.Loader)
		fmt.Printf("Số lượng trong giỏ hàng: %d mods\n", len(Cart))
		fmt.Println("--------------------------------------------")
		fmt.Println("🔥 MOD PHỔ BIẾN (Tải nhanh):")
		if len(popularHits) == 0 {
			fmt.Println("(Không tải được danh sách phổ biến hoặc không có kết quả phù hợp)")
		} else {
			for i, hit := range popularHits {
				fmt.Printf("[%d] %s\n    -> %s\n", i+1, hit.Title, hit.Description)
			}
		}
		fmt.Println("--------------------------------------------")
		fmt.Println("[s] Tìm kiếm mod khác theo tên")
		fmt.Println("[0] Quay lại Menu chính")
		fmt.Println()
		fmt.Print("Chọn số để thêm mod phổ biến, hoặc nhập 's' để tìm kiếm: ")

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
			fmt.Printf("\nĐã thêm [%s] vào giỏ hàng!\n", selected.Title)
			Pause()
		}
	}
}

// Màn hình tìm kiếm chi tiết
func SearchScreen(packName string, cfg VersionConfig) {
	reader := bufio.NewReader(os.Stdin)
	for {
		Header()
		fmt.Printf("TÌM KIẾM MOD cho [ %s ] (MC: %s | Loader: %s)\n", packName, cfg.MinecraftVersion, cfg.Loader)
		fmt.Printf("Giỏ hàng: %d mods\n", len(Cart))
		fmt.Println("--------------------------------------------")
		fmt.Print("Nhập tên mod cần tìm (hoặc gõ 'back' để về lại danh sách phổ biến): ")
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

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			fmt.Println("Lỗi tạo request:", err)
			Pause()
			continue
		}
		req.Header.Set("User-Agent", "VerityApp/ModsIn/1.0 (contact@verity.gg)")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Lỗi kết nối Modrinth API:", err)
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
			fmt.Println("\nKhông tìm thấy mod nào phù hợp.")
			Pause()
			continue
		}

		for {
			Header()
			fmt.Printf("KẾT QUẢ TÌM KIẾM CHO: '%s'\n", query)
			fmt.Printf("Giỏ hàng: %d mods\n", len(Cart))
			fmt.Println("--------------------------------------------")
			for i, hit := range result.Hits {
				fmt.Printf("[%d] %s\n    -> %s\n", i+1, hit.Title, hit.Description)
			}
			fmt.Println("--------------------------------------------")
			fmt.Println("[0] Tìm từ khóa khác")
			fmt.Println()
			fmt.Print("Chọn số để THÊM VÀO GIỎ HÀNG (0 để đổi từ khóa): ")

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
				fmt.Printf("\nĐã thêm [%s] vào giỏ hàng!\n", selected.Title)
				Pause()
			}
		}
	}
}

func ViewCartAndDownload() {
	Header()
	fmt.Println("=== GIỎ HÀNG MOD CHỜ TẢI ===")
	if len(Cart) == 0 {
		fmt.Println("Giỏ hàng đang trống!")
		Pause()
		return
	}

	for i, item := range Cart {
		fmt.Printf("%d. %s\n", i+1, item.ProjectName)
	}

	fmt.Println()
	fmt.Print("Chọn Modpack muốn lưu tất cả các mod này vào: ")
	
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
		fmt.Println("Modpack không tồn tại.")
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

	fmt.Println("\nBắt đầu tiến trình tải xuống hàng loạt...")

	downloadQueue := make([]CartItem, len(Cart))
	copy(downloadQueue, Cart)
	processedIDs := make(map[string]bool)

	for i := 0; i < len(downloadQueue); i++ {
		item := downloadQueue[i]
		if processedIDs[item.ProjectID] {
			continue
		}
		processedIDs[item.ProjectID] = true

		fmt.Printf("\nĐang xử lý tải: %s ...\n", item.ProjectName)
		
		depIDs := DownloadModVersionAndGetDeps(item.ProjectID, targetPack, cfg)
		
		for _, depID := range depIDs {
			if !processedIDs[depID] {
				downloadQueue = append(downloadQueue, CartItem{ProjectID: depID, ProjectName: "Thư viện phụ thuộc (" + depID + ")"})
			}
		}
	}

	fmt.Println("\nĐã tải xong toàn bộ giỏ hàng thành công!")
	Cart = nil
	Pause()
}

func DownloadModVersionAndGetDeps(projectID string, packName string, cfg VersionConfig) []string {
	versionsURL := fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", projectID)
	resp, err := http.Get(versionsURL)
	if err != nil {
		fmt.Println("Lỗi lấy thông tin version:", err)
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

	// Ưu tiên 1: Bản Release mới nhất
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

	// Ưu tiên 2: Nếu không có release, tìm bản Beta/Alpha mới nhất
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
		fmt.Println(" -> Không tìm thấy phiên bản tương thích cho MC/Loader này.")
		return nil
	}

	outPath := filepath.Join(PACKS, packName, targetFileName)
	out, err := os.Create(outPath)
	if err != nil {
		fmt.Println(" -> Lỗi tạo file:", err)
		return nil
	}
	defer out.Close()

	fileResp, err := http.Get(targetFileUrl)
	if err != nil {
		fmt.Println(" -> Lỗi tải file từ server:", err)
		return nil
	}
	defer fileResp.Body.Close()

	io.Copy(out, fileResp.Body)
	fmt.Printf(" -> Đã tải xong: %s\n", targetFileName)

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
