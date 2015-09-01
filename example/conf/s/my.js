function form_sleep_submit(){
	var n=findByName(this,'n');
	console && console.log("form_sleep:",n.val())
}

function form_sleep_jsonp(data){
	alert(data.data)
}