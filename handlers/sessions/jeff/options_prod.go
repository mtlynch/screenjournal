//go:build !dev

package jeff

import "github.com/abraithwaite/jeff"

func extraOptions() []func(*jeff.Jeff) {
	return []func(*jeff.Jeff){}
}
