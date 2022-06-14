package dnetserver

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func saveState(pool string) {
	stateData, err := json.Marshal(Pools[pool].Info)
	if err != nil {
		log.Fatalf("Couldn't marshal stateData: %v", err)
	}

	newStateFilePath := path.Join(TempDirectoryPath, pool)
	stateFilePath := path.Join(StateDirectoryPath, pool)

	if err = ioutil.WriteFile(newStateFilePath, stateData, 0600); err != nil {
		log.Fatalf("Failed to write state data to state file: %v", err)
	}

	if err = os.Rename(newStateFilePath, stateFilePath); err != nil {
		log.Fatalf("Failed to write state data to state file: %v", err)
	}
}

// LoadState loads the state from the statefiles in the state directory
func LoadState() {
	if _, err := os.Stat(StateDirectoryPath); err != nil {
		if os.IsNotExist(err) {
			log.Print("Tried loading state from state directory but it's non-existent")
			return
		}
		log.Fatalf("Couldn't stat state directory: %v", err)
	}

	files, err := ioutil.ReadDir(StateDirectoryPath)
	if err != nil {
		log.Fatalf("Couldn't read state directory: %v", err)
	}

	for i := 0; i < len(files); i++ {
		stateFilePath := path.Join(StateDirectoryPath, files[i].Name())
		fileData, err := ioutil.ReadFile(stateFilePath)
		if err != nil {
			log.Fatalf("Couldn't read state file %s: %v", stateFilePath, err)
		}

		Pools[files[i].Name()] = &Pool{
			Info: PoolInfo{},
		}

		err = json.Unmarshal(fileData, &Pools[files[i].Name()].Info)
		if err != nil {
			log.Fatalf("Failed to un-marshal state file %s: %v", stateFilePath, err)
		}
	}
}
