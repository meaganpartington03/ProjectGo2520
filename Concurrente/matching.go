//Partie 2 -
//Meagan Partington - 300416906
//Anastasia Sardovskyy- 300426037

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The Resident data type
type Resident struct {
	residentID     int
	firstname      string
	lastname       string
	rol            []string // resident rank order list
	matchedProgram string   // will be "" for unmatched resident
	nextOffer      int      //index du prochain programme a contacter dans ROL
}

// The Program data type
type Program struct {
	programID         string
	name              string
	nPositions        int   // number of positions available (quota)
	rol               []int // program rank order list
	selectedResidents []int //TO ADD : liste des IDs des residents selectionner
}

// Parse a resident's ROL
func parseRol(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}

// Parse a program's ROL
func parseIntRol(s string) []int {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return []int{}
	}
	parts := strings.Split(s, ",")
	var ints []int
	for _, part := range parts {
		pid, _ := strconv.Atoi(strings.TrimSpace(part))
		ints = append(ints, pid)
	}
	return ints
}

// ReadCSV reads a CSV file into a map of Resident
func ReadResidentsCSV(filename string) (map[int]*Resident, error) {

	// map to store residents by ID
	residents := make(map[int]*Resident)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	// Skip header if present (assuming it is)
	for i, record := range records {
		if i == 0 && record[0] == "id" {
			continue
		}
		if len(record) < 4 {
			return nil, fmt.Errorf("invalid record at line %d: %v", i+1, record)
		}

		// Parse ID
		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("invalid ID at line %d: %w", i+1, err)
		}

		if _, exists := residents[id]; exists {
			fmt.Println(id)
		}

		residents[id] = &Resident{
			residentID:     id,
			firstname:      record[1],
			lastname:       record[2],
			rol:            parseRol(record[3]),
			matchedProgram: "",
		}
	}

	return residents, nil
}

// reads a CSV file into a map of Program
func ReadProgramsCSV(filename string) (map[string]*Program, error) {

	// map to store programs by ID
	programs := make(map[string]*Program)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	// Skip header if present (assuming it is)
	for i, record := range records {
		if i == 0 && record[0] == "id" {
			continue
		}
		if len(record) < 4 {
			return nil, fmt.Errorf("invalid record at line %d: %v", i+1, record)
		}

		// Parse number of positions
		np, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, fmt.Errorf("invalid number at line %d: %w", i+1, err)
		}

		programs[record[0]] = &Program{
			programID:  record[0],
			name:       record[1],
			nPositions: np,
			rol:        parseIntRol(record[3]),
		}

	}

	return programs, nil
}

// getRank return la position d'un resident dans la ROL d'un programme
func getRank(rol []int, rid int) int {
	for i, id := range rol {
		if id == rid { //si on trouve le resident dans la liste, return son rang
			return i
		}
	}
	return len(rol) //le resident n'est pas dans la liste (aka rang tres bas donc non preferer)
}

// algo McVittie-Wilson version concurrent
func offerCon(wg *sync.WaitGroup, mu *sync.Mutex, rid int, residents map[int]*Resident, programs map[string]*Program) {
	resident := residents[rid] //obtenir le resident depuis le map

	wg.Add(1)
	go func() {
		defer wg.Done()

		mu.Lock()
		if resident.nextOffer >= len(resident.rol) {
			mu.Unlock()
			return
		}
		pid := resident.rol[resident.nextOffer]
		resident.nextOffer++
		mu.Unlock()

		evaluateCon(wg, mu, rid, pid, residents, programs)
	}()
}

func evaluateCon(wg *sync.WaitGroup, mu *sync.Mutex, rid int, pid string, residents map[int]*Resident, programs map[string]*Program) {
	program := programs[pid] //obtenir le programme du map

	wg.Add(1)

	go func() {
		defer wg.Done()

		mu.Lock()

		if len(program.selectedResidents) < program.nPositions {
			program.selectedResidents = append(program.selectedResidents, rid)
			residents[rid].matchedProgram = pid

		} else {
			worstRank := -1
			worstRid := -1

			for _, currentRid := range program.selectedResidents {
				rank := getRank(program.rol, currentRid)
				if rank > worstRank {
					worstRank = rank
					worstRid = currentRid
				}
			}

			newRank := getRank(program.rol, rid)
			if newRank < worstRank {
				for i, id := range program.selectedResidents {
					if id == worstRid {
						program.selectedResidents[i] = rid
						break
					}
				}

				residents[rid].matchedProgram = pid
				residents[worstRid].matchedProgram = ""
				offerCon(wg, mu, worstRid, residents, programs)

			} else {
				offerCon(wg, mu, rid, residents, programs)
			}
		}
		mu.Unlock()
	}()
}

func printResults(residents map[int]*Resident, programs map[string]*Program) {
	//structure pour stocker une ligne de resultat
	type result struct {
		lastname    string
		firstname   string
		residentID  int
		programID   string
		programName string
	}

	var results []result //liste de toutes les lignes de resultats
	unmatchedCount := 0  //compteur de residents non matched

	//parcourir tous les residents pour construire les lignes
	for _, r := range residents {
		pid := r.matchedProgram
		pname := "NOT_MATCHED" //valeur par defaut si le resident n'est pas matched
		if pid == "" {
			//resident non matched : utiliser "XXX" comme identifiant de programme
			pid = "XXX"
			unmatchedCount++
		} else {
			//obtenir le nom du programme depuis la map
			pname = programs[pid].name
		}
		//ajouter le resultat a la liste
		results = append(results, result{r.lastname, r.firstname, r.residentID, pid, pname})
	}

	//trier les resultats par nom de famille en ordre alphab
	sort.Slice(results, func(i, j int) bool {
		return results[i].lastname < results[j].lastname
	})

	fmt.Println("lastname,firstname,residentID,programID,name")
	//afficher chaque ligne de resultat
	for _, res := range results {
		fmt.Printf("%s,%s,%d,%s,%s\n", res.lastname, res.firstname, res.residentID, res.programID, res.programName)
	}

	//calculer le nombre total de places non remplies dans tous les programmes
	availablePositions := 0
	for _, p := range programs {
		availablePositions += p.nPositions - len(p.selectedResidents)
	}

	//afficher les statistiques finales
	fmt.Printf("Number of unmatched residents: %d\n", unmatchedCount)
	fmt.Printf("Number of positions available: %d\n", availablePositions)
}

// Example usage
func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// read residents
	residents, err := ReadResidentsCSV("residentSmall.csv")
	//residents, err := ReadResidentsCSV("residentsLarge.csv")
	//residents, err := ReadResidentsCSV("residents4000.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	//read program
	programs, err := ReadProgramsCSV("programSmall.csv")
	//programs, err := ReadProgramsCSV("programsLarge.csv")
	//programs, err := ReadProgramsCSV("programs4000.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	start := time.Now()

	for id := range residents { //appeler offer pour cahque resident pour lancer l'algo
		offerCon(&wg, &mu, id, residents, programs)
	}

	end := time.Now()

	printResults(residents, programs) //afficher dans le format
	fmt.Printf("\nExecution time: %s\n", end.Sub(start))
}
