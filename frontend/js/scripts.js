const URL_SERVER = 'http://192.168.1.44:8080';


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
function setTitleLog(name, id) {
    $('title').html(id + "-Логгирование");
    $('#title').html(id + "-Логгирование");
}
function stopDeviceView() {

}
function stopLogView() {

}
function addTableHeadForLog() {
    var headers = [];
    headers.push("<tr>");
    headers.push("<th>time</th>");
    headers.push("<th>message</th>");
    headers.push("<th>from-to</th>");
    headers.push("</tr>");
    $('#main_table_head').html(headers.join(""));
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
    // getCurrentDevice(currentDevice)
}
function getCurrentDevice(id) {
    stopLogView();
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
    // startUpdatingDevice();
}
function getCurrentLog(id) {
    stopDeviceView()
    setTitleLog(devices[id], id);
    addTableHeadForLog();
    $.getJSON(urlCurrentLog + id, function (data) {
        var items = [];
        $.each(data.ls, function () {
            result = "<tr>"
            time = this['t']
            time = time.substring(11, 19)
            result += "<td>" + time + "</td>"
            result += "<td>" + this['m'] + "</td>"
            result += "<td>" + this['d'] + "</td>"
            result += "<tr>"
            items.push(result);
        });
        // items.sort();
        $('#main_table_body').html(items.join(""));
    });
    currentDevice = id
    // startUpdatingDevice();
}
// function startUpdatingDevice() {
//     if (currentDevice != 0) {
//         updateValuesSubsystem();
//         updateIntervalId = setInterval(updateValuesSubsystem, timeInterval);
//     }
// }

$(document).ready(function () {
    showDevices();

    $("#nav-left-devices").on('click', '.nave-item a', function () {
        getCurrentDevice($(this).attr('id'));
        // getCurrentLog($(this).attr('id'));
    });

    $("#get-device").click(function () {
        getCurrentDevice(currentDevice);
    });
    $("#log-device").click(function () {
        getCurrentLog(currentDevice);
    });

    $("#content-filter").on("input", function () {
        contentFilter = this.value;
        // updateContent();
    });
});
