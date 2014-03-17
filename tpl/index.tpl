<!DOCTYPE html>
<html>
	<head><title>{{.Title}} Status</title>		
		<style>
			html,
			body {
			  margin: 10px;
			  margin-top: 20px;
			}
		table, td, th
		{
			border-collapse:collapse;
			border:1px solid green;
		}
		th
		{
			background-color:#ADF6FD;			
		}
		
		td
		{
			text-align:center;
			background-color:#F0F6FD;
		}
		</style>
		<script>
		var timeId;
		timeId = setInterval("checkData()",1000);
		function checkData(){
			var time = new Date();
			if(time.getSeconds() % 30 == 0){
				clearInterval(timeId);
				document.location.reload();
			}
		}
		</script>
	</head>
	<body onload="checkData()">
	<div width="100%">
	<h1>Wellcome to GoHAProxy Monitors</h1>
	</div>
	<table width="100%">
		<thead>
			<tr>
				<th>Name</th> 
				<th>Listen Port</th> 				
				<th>Mode</th>
				<th>Type</th>								
				<th>Counter</th>
				<th>Connection</th>
				<th>Destination Address</th> 
				<th>Health</th>
			</tr>
		</thead>

		<tbody>
		{{range $key, $proxy := .ProxyStatus}}			
				{{range $pk, $dst := $proxy.DstList}}
					<tr>
					<td>
						{{$proxy.Name}}
					</td>
					<td>
						{{if eq $proxy.Mode "health"}}												
							N/A
						{{else}}
							{{$proxy.SrcPort}}
						{{end}}
					</td>
					<td>
						{{$proxy.Mode}}
					</td>
					<td>
						{{$proxy.Type}}
					</td>
					<td>
						{{$dst.Counter}}
					</td>
					<td>
						{{$dst.Connections}}
					</td>
					<td>
						{{$dst.Dst}}:{{$dst.DstPort}}
					</td>	
					<td>
						{{$dst.Health}}
					</td>				
					<tr>
				{{else}}
						<tr><td>No proxy node</td></tr>
				{{end}}
		{{else}}
			<tr><td>No proxy records.</td></tr>
		{{end}}
		</tbody>
	</table>
	

	</body>
</html>