$(document).ready(function () {
	var menu = $('#menu');
	var serverTab = menu.children().eq(0);
	var clientTab = menu.children().eq(1);
	var server = $('#server');
	var client = $('#client');

	clientTab.on("click", function () {
		clientTab.addClass("tab-active");
		serverTab.removeClass("tab-active");
		server.hide();
		client.show();
	});

	serverTab.on("click", function () {
		serverTab.addClass("tab-active");
		clientTab.removeClass("tab-active");
		client.hide();
		server.show();
	});
});
