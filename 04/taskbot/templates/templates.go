package templates

const (
	TASKS = `{{range .Tasks}}{{.ID}}. {{.Text}} by @{{.Owner}}
{{ if eq .Assignee "" }}/assign_{{.ID}}{{ else if eq .Assignee $.User }}assignee: я
/unassign_{{.ID}} /resolve_{{.ID}}{{ else }}assignee: @{{.Assignee}}{{ end }}

{{end}}`
	NEWTASK    = `Задача "{{.Text}}" создана, id={{.ID}}`
	ASSIGNTASK = `Задача "{{.Text}}" назначена на {{ if eq .Assignee .User }}вас{{ else }}@{{.Assignee}}{{ end }}`
	UNASSIGN   = `Задача "{{.Text}}" осталась без исполнителя`
	RESOLVE    = `Задача "{{.Text}}" выполнена{{ if ne .Receiver .User }} @{{.User}}{{ end }}`
	MYTASKS    = `{{range .Tasks}}{{.ID}}. {{.Text}} by @{{.Owner}}
{{ if eq .Assignee "" }}/assign_{{.ID}}{{ else }}/unassign_{{.ID}} /resolve_{{.ID}}{{ end }}

{{end}}`
	OWNERTASKS = `{{range .Tasks}}{{.ID}}. {{.Text}} by @{{.Owner}}
{{ if eq .Assignee "" }}/assign_{{.ID}}{{ else if eq .Assignee $.User }}/unassign_{{.ID}} /resolve_{{.ID}}{{ end }}

{{end}}`
	HELP = `/tasks - показывает весь список задач
/new XXX YYY ZZZ - создаёт новую задачу
/assign_$ID - делает пользователя исполнителем задачи
/unassign_$ID - снимает задачу с текущего исполнителя
/resolve_$ID - выполняет задачу, удаляет её из списка
/my - показывает задачи, которые назначены на меня
/owner - показывает задачи, которые были созданы мной`
)
