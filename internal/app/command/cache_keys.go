package command

import "fmt"

func tasksCacheKey(teamID string) string {
	return fmt.Sprintf("tasks:team:%s", teamID)
}
