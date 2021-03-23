function getData() {
	const query = new URLSearchParams(window.location.search);
	var key = query.get("key");
	if (!key) {
		key = prompt("Enter key");
	}
	const resource = query.get("resource");
	const linkEl = document.querySelector("#link");
	const linkhref = window.location.protocol + "//" +window.location.host + "/" + resource;
	linkEl.innerHTML = linkhref;
	linkEl.href = linkhref;
	fetch("/"+resource+"?key="+key)
		.then(response => {
			if (response.status != 200) {
				return;
			}
			return response.text();
		}).then(text => {
			const lines = text.split("\n");
			const headerSplit = lines[0].split(" ");
			const resourceEl = document.querySelector("#resource");
			resourceEl.innerHTML = headerSplit[0];
			resourceEl.href = headerSplit[0];
			document.querySelector("#key").innerHTML = headerSplit[1];
			document.querySelector("#expires").innerHTML = headerSplit[2];
			const visitorList = document.querySelector("#visitor-list");
			for (var i = 1; i < lines.length; i++) {
				var p = document.createElement("p");
				p.classList.add("visitor");
				p.innerHTML = lines[i];
				visitorList.appendChild(p);
			}
		});
}