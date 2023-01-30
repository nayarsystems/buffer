package vars

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_String(t *testing.T) {
	vb := CreateVarsBank()
	varname := "V_STR"
	vb.InitVar(varname, "Hello World!", nil)

	v, err := vb.Get(varname)
	require.Nil(t, err)
	require.Equal(t, v, "Hello World!")

	err = vb.Set(varname, "Hello!")
	require.Nil(t, err)

	v, err = vb.Get(varname)
	require.Nil(t, err)
	require.Equal(t, v, "Hello!")
}
