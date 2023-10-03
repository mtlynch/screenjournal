//go:build !dev

package jeff

import "github.com/mtlynch/jeff"

func extraOptions() []func(*jeff.Jeff) {
	return []func(*jeff.Jeff){}
}
