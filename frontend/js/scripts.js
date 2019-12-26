var URL_SERVER = 'http://192.168.1.16:8080';


var urlList = URL_SERVER + '/list';
var urlCurrentController = URL_SERVER + '/device?id=';
var urlCurrentLog = URL_SERVER + '/l?id=';
var devices = [];

var defaultDevice = 10001;
var currentDevice = defaultDevice;
var timeInterval = 500;
var updateIntervalId = 0;

function setTitleDevice(name, id) {
    $('title').html(id + "-" + name);
    $('#title').html(id + "-" + name);
}
function addTableHeadForDevice() {
    var headers = [];
    headers.push("<tr>");
    headers.push("<th>description</th>");
    headers.push("<th>value</th>");
    headers.push("</tr>");
    $('#main_table_head').html(headers.join(""));

}
function showDevices() {
    urlList = window.location.origin + '/list';
    $.getJSON(urlList, function (data) {
        var items = [];
        devices = [];

        $.each(data.devs, function () {
            devices[this['id']] = this['name'];
            items.push("<li class='nave-item'> <a class='nav-link' href='#' id='" + this['id'] + "'  >" + this['id'] + "</a></li>");
        });

        items.sort();

        $("#nav-left-devices").html(items.join(""));

    });
}
function getCurrentDevice(id) {
    urlCurrentController = window.location.origin + '/device?id=';
    setTitleDevice(devices[id], id);
    addTableHeadForDevice();
    $.getJSON(urlCurrentController + id, function (data) {
        var items = [];
        $.each(data.dev, function () {
            result = "<tr>"
            result += "<td>" + this['desc'] + "</td>"
            result += "<td>" + this['value'] + "</td>"
            result += "<tr>"
            items.push(result);
        });
        // items.sort();
        $('#main_table_body').html(items.join(""));
    });
    currentDevice = id
}

$(document).ready(function () {
    showDevices();

    $("#nav-left-devices").on('click', '.nave-item a', function () {
        getCurrentDevice($(this).attr('id'));
    });

    $("#get-device").click(function () {
        getCurrentDevice(currentDevice);
    });
    $("#log-device").click(function () {
        getCurrentLog(currentDevice);
    });

    $("#content-filter").on("input", function () {
        contentFilter = this.value;
    });
});
