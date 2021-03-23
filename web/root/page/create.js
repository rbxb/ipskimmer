const resourceInput = document.querySelector("#resource-input");
const submitButton = document.querySelector("#submit-button");

function createLink() {
	try {
		const value = resourceInput.value;
		fetch("/sk-create?resource="+btoa(value))
			.then(response => {
				if (response.status != 200) {
					return;
				}
				return response.text();
			}).then(text => {
				const split = text.split(" ");
				window.location.href = "/page/view?resource=" + split[0] + "&key=" + split[1];
			});
	} catch(err) {
		console.error(err);
	}
}

function inputChange() {
	allowSubmit();
	const value = resourceInput.value;
	if (value.length == 0) {
		preventSubmit();
		return;
	}
	var u;
	try {
		u = new URL(value);
	} catch(err) {
		preventSubmit();
		return;
	}
}

function preventSubmit() {
	submitButton.disabled = true;
}

function allowSubmit() {
	submitButton.disabled = false;
}