package main
import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"io"
	"os/exec"
)
var (
	ROOT    string
	PACKS   string
	MC      string
	COUNT   int
	MODS    int
	CHOICE  string
	NAME    string
	CHOOSE  string
	TARGET  string
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
		fmt.Println("0. Exit")
		fmt.Println()
		fmt.Print("Select: ")
		reader := bufio.NewReader(os.Stdin)
		CHOICE, _ = reader.ReadString('\n')
		CHOICE = strings.TrimSpace(CHOICE)
		if CHOICE == "1" {
			Create()
			continue
		}
		if CHOICE == "2" {
			List()
			continue
		}
		if CHOICE == "3" {
			Switch()
			continue
		}
		if CHOICE == "0" {
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
	fmt.Println("              ModsIn v1.0")
	fmt.Println("       Minecraft Modpack Manager")
	fmt.Println("============================================")
	fmt.Println()
}
func Create() {
	Header()
	fmt.Println("Create Modpack")
	fmt.Println()
	fmt.Print("Enter Name: ")
	reader := bufio.NewReader(os.Stdin)
	NAME, _ = reader.ReadString('\n')
	NAME = strings.TrimSpace(NAME)
	if NAME == "" {
		return
	}
	if _, err := os.Stat(filepath.Join(PACKS, NAME)); err == nil {
		fmt.Println()
		fmt.Println("Modpack already exists.")
		Pause()
		return
	}
	os.Mkdir(filepath.Join(PACKS, NAME), os.ModePerm)
	fmt.Println()
	fmt.Println("Created:")
	fmt.Println(filepath.Join(PACKS, NAME))
	fmt.Println()
	fmt.Println("Place your .jar files into this folder.")
	Pause()
}
func List() {
	Header()
	fmt.Println("List Modpacks")
	fmt.Println()
	dirs, _ := os.ReadDir(PACKS)
	for _, dir := range dirs {
		if dir.IsDir() {
			fmt.Println(dir.Name())
		}
	}
	fmt.Println()
	Pause()
}
func Switch() {
	Header()
	fmt.Println("List Modpacks")
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
	fmt.Println("[Y] Yes")
	fmt.Println("[N] No")
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
var DUPLICATE string
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
