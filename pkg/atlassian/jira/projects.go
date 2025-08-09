package jira

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

func (c *client) GetProjects() {
	project, _, err := c.Client.Project.Get(context.Background(), "JAR", nil)
	if err != nil {
		slog.Error("Failed to get project", "error", err)
		os.Exit(1)
	}
	fmt.Println(project.Key)

}
