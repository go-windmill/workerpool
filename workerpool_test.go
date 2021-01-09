package workerpool

import (
	"fmt"
	"testing"
)

func Test_New_Returned(t *testing.T) {
	t.Parallel()

	p := New("test", 2)

	p.Add(func() {
		fmt.Println("Handling...")
	})

	p.CloseAndWait()
}
