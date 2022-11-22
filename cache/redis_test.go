package cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedisClient(t *testing.T) {
	//given
	repo := NewRedisClient()

	//when
	pong, err := repo.client.Ping().Result()

	//then
	require.Nil(t, err)
	require.Equal(t, pong, "PONG")
}

func TestRedisInit(t *testing.T) {

	//given
	repo := NewRedisClient()
	repo.InitRepository()

	//when
	err := repo.AddDeny("src1", 1, func(v string) (string, error) {
		return "foo", nil
	})
	require.Nil(t, err)

	//then
	fmt.Println(repo.FindAllDenyList())
	val, err := repo.GetRepoSize()
	require.Nil(t, err)
	require.Equal(t, val, 1)

	err = repo.InitRepository()
	require.Nil(t, err)

	val, err = repo.GetRepoSize()
	require.Nil(t, err)
	require.Equal(t, val, 0)

	fmt.Println(repo.FindAllDenyList())

}
