 var idChartUpdater = '';
var periodChartUpdating = 1000;

var myChart;
var isPause = false;
var chartViewData;

var configChart = {
    type: 'line',
    data: {
        labels: [],
        datasets: []
    },
    options: {
        responsive: true,
        scales: {
            yAxes: [{
                ticks: {
                    beginAtZero: false
                }
            }]
        }
    }
};

function getHash(source) {
        var hash = 0, i, chr;
        for (i = 0; i < source.length; i++) {
          chr   = source.charCodeAt(i);
          hash  = ((hash << 5) - hash) + chr;
          hash |= 0; // Convert to 32bit integer
        }
        return hash;
      
}

function getCurrentTime() {
    var today = new Date();
    var seconds = today.getSeconds();
    seconds = (seconds < 10 ? '0' : '') + seconds;
    var result = today.getMinutes() + ":" +  seconds;
    return result;
}

function getValue(input) {
    if ( Array.isArray(input) ) {
        return input[0];
    }
    else {
        if (input[input.length-1] == ' ') {
            return input.slice(0, -1);
        }
        else {
            return input;
        }
    }
}

function showValue(indx) {
    $.getJSON( chartViewData.url + chartViewData.data[indx]['name'], function(data){
        for ( var j=0; j < data['values'].length; j++  ) {
            if (data['values'][j]['name'] == chartViewData.data[indx]['value'] ) {
                var inputValue = getValue(data['values'][j]['value']);
                var currentValue;
                if ( inputValue === 'false') { currentValue =0;}
                else if (inputValue === 'true') { currentValue = 1;}
                else { currentValue = inputValue; }

                configChart.data.datasets[indx].data.push(currentValue);
                myChart.update();

                var idValue = "#value" + getHash(chartViewData.data[indx]['name'] + ":" + chartViewData.data[indx]['value']);
                $(idValue).html(currentValue);
                break;
            }
        }
    });
}

function chartUpdating() {
    configChart.data.labels.push(getCurrentTime());
    for ( var i=0; i < configChart.data.datasets.length; i++) {
        showValue(i);
    }
}

function pauseChartUpdating() {
    if (idChartUpdater != '') {
        stopChartUpdating();
    }
    else {
        startChartUpdating();
    }
}

function startChartUpdating() {
    if (idChartUpdater == '') {
        chartUpdating();
        idChartUpdater = setInterval( chartUpdating, periodChartUpdating);
    }
}

function stopChartUpdating() {
    if (idChartUpdater != '') {
        clearInterval(idChartUpdater);
        idChartUpdater = '';
    }
}

var chartColors = [
    'rgb(255, 99, 132)',
    'rgb(153, 102, 255)',
    'rgb(255, 159, 64)',
    'rgb(255, 205, 86)',
    'rgb(55, 199, 32)',
    'rgb(25, 29, 132)',
    'rgb(255, 199, 132)',
    'rgb(75, 192, 192)',
    'rgb(54, 162, 235)',
    'rgb(201, 203, 207)'
];

function showTitle() {
    var title = "";
    for (var i=0; i < chartViewData.data.length; i++ ) {
        if (title.length > 0) {
            title += ", ";
        }
        title += chartViewData.data[i]['value'];
    }
    $('title').html(title);
}

function showChart() {
    var ctx = document.getElementById('myChart')
    myChart = new Chart(ctx, configChart);

    for (var i=0; i < chartViewData.data.length; i++ ) {
		var newColor = chartColors[i];
        var ds = {
            label: chartViewData.data[i]['name'] + ":" + chartViewData.data[i]['value'],
            lineTension: 0,
            backgroundColor: newColor,
            borderColor: newColor,
            data: [],
            fill: false
        };
        configChart.data.datasets.push(ds);
    }
    myChart.update();
}

function showTableValues() {
    var out = "<table>";
    for (var i=0; i < chartViewData.data.length; i++ ) {
        out += "<tr><td>"+ chartViewData.data[i]['name'] + ":" + chartViewData.data[i]['value']+": </td><td id='value" + getHash(chartViewData.data[i]['name'] + ":" + chartViewData.data[i]['value']) + "'></td></tr>";
    }
    out += "</table>"
    $("#values").html(out);    
}

$(document).ready(function(){
    const urlParams = new URLSearchParams(window.location.search);
    const data = urlParams.get('data');
    chartViewData = JSON.parse(data);

    showTitle();

    showChart();
    startChartUpdating();
    isPause = false;

    showTableValues();

    $('#btn-pause').click(function() {
        if ( !isPause ) {
            stopChartUpdating();
            $(this).addClass('active');
            isPause = true;
        }
        else {
            startChartUpdating();
            isPause = false;
            $(this).removeClass('active');
        }
    });
});