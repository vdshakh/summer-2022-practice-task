package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const timeLayout = "15:04:05"

type CriteriaMap map[string]struct{}

type Trains []Train

type CustomTime time.Time

// UnmarshalJSON Parses the json string in the custom format
func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	nt, err := time.Parse(timeLayout, s)
	*ct = CustomTime(nt)

	return
}

type Train struct {
	TrainID            int       `json:"trainId"`
	DepartureStationID int       `json:"departureStationId"`
	ArrivalStationID   int       `json:"arrivalStationId"`
	Price              float32   `json:"price"`
	ArrivalTime        time.Time `json:"arrivalTime"`
	DepartureTime      time.Time `json:"departureTime"`
}

func (t *Train) UnmarshalJSON(data []byte) error {
	var aux struct {
		TrainID            int
		DepartureStationID int
		ArrivalStationID   int
		Price              float32
		ArrivalTime        CustomTime
		DepartureTime      CustomTime
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&aux); err != nil {
		return fmt.Errorf("decode train: %v", err)
	}

	t.TrainID = aux.TrainID
	t.DepartureStationID = aux.DepartureStationID
	t.ArrivalStationID = aux.ArrivalStationID
	t.Price = aux.Price
	t.ArrivalTime = time.Time(aux.ArrivalTime)
	t.DepartureTime = time.Time(aux.DepartureTime)

	return nil
}

// String returns the train in the custom format
func (t Train) String() string {
	output, _ := fmt.Printf("TrainID \t DepartureStationID \t\t ArrivalStationID \t Price \t\t\t "+
		"ArrivalTime \t\t DepartureTime \n %v\t\t %v\t\t\t\t %v\t\t\t %v\t\t\t %v\t\t %v", t.TrainID,
		t.DepartureStationID, t.ArrivalStationID, t.Price, t.ArrivalTime.Format("15:04:05"),
		t.DepartureTime.Format("15:04:05"))

	return string(output)
}

var (
	criteriaErr          = errors.New("unsupported criteria")
	emptyDepartureErr    = errors.New("empty departure station")
	emptyArrivalErr      = errors.New("empty arrival station")
	badArrivalInputErr   = errors.New("bad arrival station input")
	badDepartureInputErr = errors.New("bad arrival departure input")
)

var validCriteria = CriteriaMap{
	"price":          {},
	"arrival-time":   {},
	"departure-time": {},
}

func main() {
	depStation, arrStation, criteria := input()

	result, err := FindTrains(depStation, arrStation, criteria)
	if err != nil {
		fmt.Printf("\nfindTrains failed: %w", err)
	}

	PrintTrains(result)
}

func FindTrains(departureStation, arrivalStation, criteria string) (Trains, error) {
	err := validator(departureStation, arrivalStation, criteria)
	if err != nil {
		return nil, fmt.Errorf("validator failed: %w", err)
	}

	affordableTrains, err := SelectTrains(departureStation, arrivalStation)
	if err != nil {
		return nil, fmt.Errorf("selectTrains failed: %w", err)
	}

	if len(affordableTrains) <= 1 {
		return affordableTrains, nil
	}

	sortedTrains := SortTrains(affordableTrains, criteria)

	if len(sortedTrains) >= 3 {
		topTrains := SeparateTopTrains(sortedTrains)

		return topTrains, nil
	}

	return sortedTrains, nil
}

func input() (depStation, arrStation, criteria string) {
	fmt.Print("Enter Departure Station: ")
	depStation, err := readInput()
	if err != nil {
		fmt.Printf("\nreadInput for departureStation failed: %w", err)
	}

	fmt.Print("Enter Arrival Station: ")
	arrStation, err = readInput()
	if err != nil {
		fmt.Printf("\nreadInput for arrivalStation failed: %w", err)
	}

	fmt.Print("Enter Criteria: ")
	criteria, err = readInput()
	if err != nil {
		fmt.Printf("\nreadInput for criteria failed: %w", err)
	}

	return depStation, arrStation, criteria
}

func readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	userInput, err := reader.ReadString('\n') // ReadString will block until the delimiter is entered
	if err != nil {
		fmt.Println("An error occured while reading input. Please try again", err)

		return "", err
	}

	userInput = strings.TrimSuffix(userInput, "\n") // remove the delimeter from the string
	fmt.Println(input)

	return userInput, nil
}

func validator(departureStation, arrivalStation, criteria string) error {
	if err := validateEmpty(departureStation); err != nil {
		return emptyDepartureErr
	}

	if err := validateEmpty(arrivalStation); err != nil {
		return emptyDepartureErr
	}

	if err := validateIsNaturalNumber(departureStation); err != nil {
		return badDepartureInputErr
	}

	if err := validateIsNaturalNumber(arrivalStation); err != nil {
		return badArrivalInputErr
	}

	if _, value := validCriteria[criteria]; !value {
		return criteriaErr
	}

	return nil
}

func validateEmpty(s string) error {
	if s == "" {
		return fmt.Errorf("value of input is empty")
	}

	return nil
}

func validateIsNaturalNumber(s string) error {
	value, _ := strconv.Atoi(s)

	if value <= 0 {
		return fmt.Errorf("value is not a natural number")
	}

	return nil
}

func importInfo() (Trains, error) {
	jsonFile, err := os.Open("data.json")
	if err != nil {
		return nil, fmt.Errorf("os.Open returns an error: %v", err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var trainSchedule []Train
	if err = json.Unmarshal(byteValue, &trainSchedule); err != nil {
		return nil, fmt.Errorf("error during Unmarshal: %v", err)
	}

	return trainSchedule, nil
}

func SelectTrains(departureStation, arrivalStation string) (Trains, error) {
	var affordableTrains Trains

	trainSchedule, err := importInfo()
	if err != nil {
		return nil, fmt.Errorf("importInfo failed: %w", err)
	}

	departure, _ := strconv.Atoi(departureStation)
	arrival, _ := strconv.Atoi(arrivalStation)

	for _, v := range trainSchedule {
		if v.DepartureStationID == departure && v.ArrivalStationID == arrival {
			affordableTrains = append(affordableTrains, v)
		}
	}

	return affordableTrains, nil
}

func SortTrains(affordableTrains Trains, criteria string) Trains {
	var sortedTrains Trains

	switch criteria {
	case "price":
		sortedTrains = SortTrainsByPrice(affordableTrains)

	case "arrival-time":
		sortedTrains = SortTrainsByArrival(affordableTrains)

	default: // aka "departure-time"
		sortedTrains = SortTrainsByDeparture(affordableTrains)
	}

	return sortedTrains
}

func SortTrainsByPrice(affordableTrains Trains) Trains {
	sort.Slice(affordableTrains, func(i, j int) bool {
		if affordableTrains[i].Price != affordableTrains[j].Price {
			return affordableTrains[i].Price < affordableTrains[j].Price
		}

		return affordableTrains[i].Price < affordableTrains[j].Price
	})

	return affordableTrains
}

func SortTrainsByArrival(affordableTrains Trains) Trains {
	sort.Slice(affordableTrains, func(i, j int) bool {
		if affordableTrains[i].ArrivalTime != affordableTrains[j].ArrivalTime {
			return affordableTrains[i].ArrivalTime.Before(affordableTrains[j].ArrivalTime)
		}

		return affordableTrains[i].ArrivalTime.Before(affordableTrains[j].ArrivalTime)
	})

	return affordableTrains
}

func SortTrainsByDeparture(affordableTrains Trains) Trains {
	sort.Slice(affordableTrains, func(i, j int) bool {
		if affordableTrains[i].DepartureTime != affordableTrains[j].DepartureTime {
			return affordableTrains[i].DepartureTime.Before(affordableTrains[j].DepartureTime)
		}

		return affordableTrains[i].DepartureTime.Before(affordableTrains[j].DepartureTime)
	})

	return affordableTrains
}

func SeparateTopTrains(affordableTrains Trains) Trains {
	var topTrains Trains

	for i := 0; i < 3; i++ {
		topTrains = append(topTrains, affordableTrains[i])
	}

	return topTrains
}

func PrintTrains(trains Trains) {
	for _, v := range trains {
		fmt.Printf("%v\n", v)
	}
}
