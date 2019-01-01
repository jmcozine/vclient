package main

const tmpl = `
<!DOCTYPE html>
<html lang="en-US">
  <head>
    <meta charset="UTF-8">
    <title>vSphere Client</title>
    <script src="http://code.jquery.com/jquery-1.8.3.min.js"></script>
    <script src="http://code.jquery.com/ui/1.8.16/jquery-ui.min.js"></script>
    <script src="/static/wmks.min.js"></script>
    <link rel="stylesheet" href="/static/app.css">
  </head>
  <body>
    {{if .Token}}
    <div id="wmksContainer" style="position:absolute;width:100%;height:100%"></div>
    <script>
      var wmks = WMKS.createWMKS("wmksContainer",{})
          .register(WMKS.CONST.Events.CONNECTION_STATE_CHANGE, function(event,data){
              if(data.state == WMKS.CONST.ConnectionState.CONNECTED){
                  console.log("connection state change : connected");}
          });
      wmks.connect("wss://{{.Host}}:443/ticket/{{.Token}}");
    </script>
    {{else if .Auth}}
    <p>Account: {{.User}}</p>
    <ul>
      {{range .VMs}}
      <li>
        {{if eq .Summary.Runtime.PowerState "poweredOn"}}
        <input type="button" value="▶" class="on">
        {{else}}
        <input type="button" value="■" class="off">
        {{end}}
        <a href="/?vm={{.Summary.Config.Name}}" target="_blank">{{.Summary.Config.Name}}</a>
      </li>
      {{end}}
    </ul>
    {{else}}
    <form method="post">
      Username:<input type="text" name="username">
      Password:<input type="password" name="password">
      <input type="submit" value="Login">
    </form>
    {{end}}
  </body>
</html>
`
