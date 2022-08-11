package main

import (
	"encoding/json"
	"fmt"
	"github.com/jcelliott/lumber"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

const Version = "1.0.1"

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}
	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		dir     string
		log     Logger
	}
)
type Options struct {
	Logger
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)
	opts := Options{}
	if options != nil {
		opts = *options
	}
	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger((lumber.INFO))
	}

	driver := Driver{
		dir:     dir,
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}
	if _, err := os.Stat(dir); err != nil {
		opts.Logger.Debug("using  '%s' (database already exists)\n", dir)
		return &driver, nil
	}
	opts.Logger.Debug("Creating the database at '%s' ... \n ", dir)
	return &driver, os.MkdirAll(dir, 0755)

}
func (d *Driver) Write(collection string, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("Missing collection - unable to save")
	}
	if resource == "" {
		return fmt.Errorf("Missing resource unable to save record")
	}
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, collection)
	fnlpath := filepath.Join(dir, resource+".json")
	tmpPath := fnlpath + ".tmp"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	b = append(b, byte('\n'))
	if err := ioutil.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, fnlpath)
}
func (d *Driver) Read(collection string, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("Missing collection - unable to read")
	}
	if resource == "" {
		return fmt.Errorf("Missing resource unable to read record (no name)")
	}
	record := filepath.Join(d.dir, collection, resource)
	if _, err := stat(record); err != nil {
		return err
	}
	b, err := ioutil.ReadFile(record + ".json")
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &v)

}
func (d *Driver) ReadAll(collection string) ([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("Missing collection - unable to read")
	}
	dir := filepath.Join(d.dir, collection)
	if _, err := stat(dir); err != nil {
		return nil, err
	}
	files, _ := ioutil.ReadDir(dir)
	var records []string

	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		records = append(records, string(b))
	}
	return records, nil

}
func (d *Driver) Delete(collection string, resource string) error {

	path := filepath.Join(collection, resource)
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.dir, path)
	fmt.Println(dir)
	switch fi, err := stat(dir); {
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file or directory named %v\n", path)

	case fi.Mode().IsDir():
		return os.RemoveAll(dir)

	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}
	return nil

}
func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]
	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}
	return m
}

func stat(path string) (fi os.FileInfo, err error) {
	fmt.Sprintln(path, "state of path")
	fmt.Sprintln(fi, "file passing not found")

	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
		fmt.Sprintln(fi, "not found")
		return fi, nil
	}
	return
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
}

func main() {
	dir := "./"
	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error", err)
	}
	employees := []User{
		{"john", "24", "233444444", "Smart Programmer",
			Address{
				"Lagos",
				"Lagos",
				"Nigeria",
				"12333",
			}},
		{"Temi", "24", "272374737", "Smart Programmer",
			Address{
				"De",
				"Edo",
				"Nigeria",
				"12333",
			}},
		{"Folake", "24", "7777777", "Greenland",
			Address{
				"Lagos",
				"Lagos",
				"Nigeria",
				"12333",
			}},
		{"Taiye", "24", "38383383", "Microsoft",
			Address{
				"De",
				"Edo",
				"Nigeria",
				"12333",
			}},
		{"Folake", "24", "7777777", "Apple",
			Address{
				"Lagos",
				"Lagos",
				"Nigeria",
				"12333",
			}},
		{"Badru", "24", "484884", "Google",
			Address{
				"De",
				"Edo",
				"Nigeria",
				"12333",
			}},
		{"Folake", "24", "737377", "Facebook",
			Address{
				"Lagos",
				"Lagos",
				"Nigeria",
				"12333",
			}},
	}

	for _, value := range employees {
		db.Write("users", value.Name, User{
			Name:    value.Name,
			Age:     value.Age,
			Address: value.Address,
			Company: value.Company,
			Contact: value.Contact,
		})
	}
	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println(records)

	allusers := []User{}

	for _, f := range records {
		employeeFound := User{}
		if err := json.Unmarshal([]byte(f), &employeeFound); err != nil {
			fmt.Println("Error", err)
		}
		allusers = append(allusers, employeeFound)
	}
	fmt.Println(allusers)

	if err := db.Delete("users", "Badru"); err != nil {
		fmt.Printf("Error", err)
	}

	if err := db.Delete("users", ""); err != nil {
		fmt.Printf("Error", err)
	}

}
