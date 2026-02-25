//Partie 2 - 
//Meagan Partington - 300416906
//Anastasia Sardovskyy


package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The Resident data type
type Resident struct {
	residentID int
	firstname string
	lastname string
	rol []string 			// resident rank order list
	matchedProgram string	// will be "" for unmatched resident
	nextOffer int  //index du prochain programme a contacter dans ROL
}

// The Program data type
type Program struct {
	programID string
	name       string
	nPositions  int 		// number of positions available (quota)
	rol []int  				// program rank order list
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
		pid,_:= strconv.Atoi(strings.TrimSpace(part))
		ints= append(ints,pid) 
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
			residentID:        id,
			firstname: record[1],
			lastname:  record[2],
			rol:     parseRol(record[3]),
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
			programID: record[0],
			name: record[1],
			nPositions:  np,
			rol:     parseIntRol(record[3]),
		}
		
	}

	return programs, nil
}

//getRank return la position d'un resident dans la ROL d'un programme
func getRank(rol []int, rid int) int {
	for i, id := range rol {
		if id == rid { //si on trouve le resident dans la liste, return son rang
			return i
		}
	}
	return len(rol) //le resident n'est pas dans la liste (aka rang tres bas donc non preferer)
}


//algo McVittie-Wilson ver sequentielle
func offer(rid int, residents map[int]*Resident, programs map[string]*Program) { //le resident rid fait un offre au prochain rpogrammde de sa ROL
	resident := residents[rid] //obtenir le resident depuis le map

	if resident.nextOffer >= len(resident.rol) { //si le resident a contacter tout ses programmes il reste
		return
	}

	pid := resident.rol[resident.nextOffer] //obetnir l'ID du prochain programme a contacter
	resident.nextOffer++ //avance l'index pour que le prochain attempt contacte le programme suivant
	evaluate(rid, pid, residents, programs) //le programme evalue l'offre du resident
}

func evaluate(rid int, pid string, residents map[int]*Resident, programs map[string]*Program) { //le programme pid evalue l'offre du resident rid
	program := programs[pid] //obtenir le programme du map

	if len(program.selectedResidents) < program.nPositions { //soit le programme a encore des places libres
		program.selectedResidents = append(program.selectedResidents, rid) //accepter le resident donc l'ajouter a la liste des selectionnes
		residents[rid].matchedProgram = pid //marque le resident comme matched au programm

	} else { //ou soit le programme est plein, ont trouve le resident le moins preferer
		worstRank := -1 //rang du pire resident actuellement selectionne
		worstRid := -1 //ID du pire resident selectionne

		for _, currentRid := range program.selectedResidents {
			rank := getRank(program.rol, currentRid) //calculet le rang de chaque resident selectionne dans le ROL du programme
			if rank > worstRank { //store le resident avec le rang le plus mauvais
				worstRank = rank
				worstRid = currentRid
			}
		}

		newRank := getRank(program.rol, rid) //calculer le rang du nouveau resdient dnas le ROL du programme
		if newRank < worstRank { //si le programme preefere le nouveau resident 
			for i, id := range program.selectedResidents { //remplacer le pire resident par le nouveau dans la liste des selectionnes
				if id == worstRid {
					program.selectedResidents[i] = rid
					break
				}
			}
			
			residents[rid].matchedProgram = pid //mettre a jour le nouveau resident accepte
			residents[worstRid].matchedProgram = "" //le resident rejeter devient libre
			offer(worstRid, residents, programs) //le resident rejeter doit iare une nouvelle offre 

		} else { 
			offer(rid, residents, programs) //le programme prefere ses residents actuels 
		}
	}
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

    // read residents
	residents, err := ReadResidentsCSV("residents.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	programs, err := ReadProgramsCSV("programs.csv") //read program
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	start := time.Now() //debut du chronometre

	for id := range residents { //appeler offer pour cahque resident pour lancer l'algo
		offer(id, residents, programs)
	}

	end := time.Now() //fin du chronometre

	printResults(residents, programs) //afficher dans le format
	fmt.Printf("\nExecution time: %s\n", end.Sub(start))
}
