$(function() {
	$('[data-toggle="popover"]').popover()
})
$("#btn_search").popover({
	"placement": "bottom",
	"html": true,
	"trigger": "manual",
	"content": "<div id='search_tip'></div>"
}).on("blur", function() {
	$(this).popover("hide");
	$("#search_keywords").popover("hide")
})
$("#search_keywords").popover({
	"placement": "bottom",
	"html": true,
	"trigger": "manual",
	"title": "搜索记录",
	"content": "<div class='list-group' id='search_result'></div>"
})
$("#btn_search").on("click", function() {
	var w = $("#search_keywords").val()
	if (w == "") {
		$("#btn_search").on("inserted.bs.popover", function() {
			$("#search_tip").html("输入关键字")
		})
		$("#btn_search").popover("show")
		return
	}
	$.get("/atom.xml", function(xml) {
		var find = ""
		$(xml).find("entry").each(function(i) {
			var c = $(this).find("content").text()
			var t = $(this).find("title").text()
			var s = $(this).find("summary").text()
			if (c.indexOf(w) >= 0 || t.indexOf(w) >= 0 || s.indexOf(w) >= 0) {
				var url = $(this).find("id").text()
				find = find + "<a href='" + url + "' class='list-group-item'>" + t + "</a>"
			}
		})
		if (find != "") {
			$("#search_keywords").on("inserted.bs.popover", function() {
				$("#search_result").html(find)
			})
			$("#search_keywords").popover("show")
		} else {
			$("#btn_search").on("inserted.bs.popover", function() {
				$("#search_tip").html("无搜索记录")
			})
			$("#btn_search").popover("show")
		}
	})
})