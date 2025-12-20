---
title: "Users from External Source"
sources:
  users:
    type: exec
    cmd: ./get-users.sh
---

# Users List

Test for `lvt-source` attribute that fetches data from an external command.

```lvt
<main lvt-source="users">
    <h2>Users</h2>

    {{if .Error}}
    <p><mark>Error: {{.Error}}</mark></p>
    {{else}}
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Name</th>
                <th>Email</th>
            </tr>
        </thead>
        <tbody>
            {{range .Data}}
            <tr data-user-id="{{.id}}">
                <td>{{.id}}</td>
                <td>{{.name}}</td>
                <td>{{.email}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{end}}

    <button lvt-click="Refresh">Refresh Data</button>
</main>
```
