package cmds
import (
	"io/ioutil"
	"fmt"
	"errors"
)
func Cat(argv []string) int {
	if len(argv) == 0 {
		// TODO: this is 'read stdin' mode
		CatUsage(errors.New("Need: file"))
		return 1
	}
	for _, fn := range argv {
		err := cat(fn)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return 0
}
func CatUsage(err error) {
	fmt.Println(err)
	fmt.Println("Usage: cat <file[ file[ file...]]>")
}
func cat(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// The function of cat is printing to stdout
	fmt.Println(string(data))
	return nil
}
