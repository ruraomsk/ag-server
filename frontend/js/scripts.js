const URL_SERVER = 'http://192.168.10.30:8080';

const Register_COIL = 0;
const Register_DI = 1;
const Register_IR = 2;
const Register_HR = 3;

const classSelectVariable = 'select-variable';

var urlSubs = URL_SERVER+'/allSubs';
var urlCurrentSubsystem = URL_SERVER+'/subinfo?name=';
var urlValuesSubsystem = URL_SERVER+'/subvalue?name=';

var urlModbuses = URL_SERVER+'/allModbuses';
var urlCurrentModbus= URL_SERVER+'/modinfo?name=';
var urlValuesModbus = URL_SERVER+'/modvalue?name=';

var urlSetSubsystemValue = URL_SERVER+'/setsubval';
var urlSetModbusValue = URL_SERVER+'/setmodval';

var subsystems = [];
var modbuses = [];

var defaultSubsystem = "AKNP1";

var currentSubsystem = "";
var currentModbus = "";

var timeInterval = 500;
var updateIntervalId = 0;

var contentFilter = '';

var selectedItems = [];
var isSelectedActive = false;

var chartData = [];

function setTitle(name) {
    $('title').html(name);
    $('#title').html(name);
}

function isContentFilter(str) {
    if ( contentFilter.length > 0  ) {
        var sample = str.toLowerCase();
        var filter = contentFilter.toLowerCase();

        var subfilter = filter.split(';');

        for (var i=0; i < subfilter.length; i++) {
            if ( sample.indexOf( subfilter[i] ) >= 0) {
                return true;
            }
        }
        return false;
    }
    return true;
}

function isSelected(variableName) {
    if ( selectedItems.length == 0 ) {
        return true;
    }

    var sample = 'select_' + variableName;
    for (var i=0; i < selectedItems.length; i++) {
        if ( selectedItems[i].toLowerCase() === sample.toLowerCase() ) {
            return true;
        }
    }
    return false;
}

function isShowSelected(nameVariable) {
    if ( isSelectedActive ) {
        if ( isSelected(nameVariable) ) {
            return true;
        }
        else {
            return false;
        }
    }
    else {
        return true;
    }
}

function clearCheckboxesById(id) {
    $(id).each( function(){
        if (this.checked) {
            this.checked = false;
        }
    });
}

function clearSelectedVariables() {
    isSelectedActive = false;
    $('#show-selected-variables').removeClass('active');
    clearCheckboxesById('.'+classSelectVariable);
    selectedItems = [];
}

function clearAllCheckboxes() {
    clearCheckboxesById('.checkboxes');
    chartData = [];
}

function clickToCheckbox(checkbox) {
    var name = checkbox.name;
    var value = checkbox.value;

    if ( checkbox.checked ) {
        var i;
        for (i=0; i < chartData.length; i++) {
            if ( chartData[i]['name'] === name && chartData[i]['value'] === value ) {
                break;
            }
        }

        if ( i == chartData.length) {
            chartData.push({name, value});
        }
    }
    else {
            for( var i = 0; i < chartData.length; i++){ 
                if ( chartData[i]['name'] === name && chartData[i]['value'] === value ) {
                    chartData.splice(i, 1); 
                    break;
                }
             }
    }
}

function startChart() {
    if (chartData.length == 0) {
        alert("Данные для построения графика не выбраны");
        return;
    }

    var currentUrl = (currentModbus != '' ) ? urlValuesModbus : urlValuesSubsystem ;
    var data = {
        'url': currentUrl,
        'data': Array.from(chartData)
    };

    var uriData = encodeURIComponent(JSON.stringify(data));
    window.open("chart.html?data="+uriData);
}

function addTableHeadForSubsystems(ips) {
    var headers = [];
    headers.push("<tr>");
    headers.push("<th>name</th>");
    headers.push("<th>desription</th>");

    for (var i = 0; i < ips.length; i++) {
        headers.push("<th>"+ ips[i]+"</th>");
    }
    headers.push("</tr>");
    $('#main_table_head').html(headers.join(""));
}

function getSlugIdSubsystem( nameSubsystem, ipSubsystem, nameVariable ) {
    //nameSubsystem+subsystems[nameSubsystem][i].replace(/\./g, '') +row['name']
    return nameSubsystem + ipSubsystem.replace(/\./g, '') + nameVariable;
}

function getRowVariable(row, nameSubsystem) {
    var result = "<tr>";
    result += "<td><input type='checkbox' class='"+classSelectVariable+"' id='select_"+row['name']+"'";
    if ( isSelected(row['name']) && selectedItems.length > 0 ) {
        result += " checked";
    }
    result += "> "+ row['name'] +"</td>";
    result += "<td>"+ row['description'] +"</td>";

    for (var i = 0; i < subsystems[nameSubsystem].length; i++) {
        result += "<td>";
        result += "<input type='checkbox' class='checkboxes form-check-inline' name='"+ nameSubsystem+":"+subsystems[nameSubsystem][i]+"' value='"+row['name']+"'>";
        result += "<span class='btn-edit editable' id='"+nameSubsystem+":"+subsystems[nameSubsystem][i]+"&name="+row['name']+"'>&#9998;</span> ";
        result += "<span id='" +getSlugIdSubsystem(nameSubsystem, subsystems[nameSubsystem][i], row['name'])+"'> </span>";
//        result += "<span class='editable' id='"+nameSubsystem+":"+subsystems[nameSubsystem][i]+"&name="+row['name']+"'><span id='" +nameSubsystem+subsystems[nameSubsystem][i].replace(/\./g, '') +row['name']+"'> </span></span>";
        result += "</td>";
    }
    result += "</tr>";
    return result;
}

function getRowRegister(row) {
    var result = "<tr>";
    result += "<td><input type='checkbox' class='"+classSelectVariable+"' id='select_"+row['name']+"'";

    if ( isSelected(row['name']) && selectedItems.length > 0 ) {
        result += " checked";
    }
    
    result += "> "+ row['name'] +"</td>";
    result += "<td>"+ row['desc'] +"</td>";

    switch(row['type']) {
        case Register_COIL:
            result += "<td>COIL</td>";
            break;
        case Register_DI:
            result += "<td>DI (ReadOnly)</td>";
            break;
        case Register_IR:
            result += "<td>IR (ReadOnly)</td>";
            break;
        case Register_HR:
            result += "<td>HR</td>";
            break;
        default:
            result += "<td>undefined</td>";
            break;
    }

    result += "<td><input type='checkbox' class='checkboxes' name='"+currentModbus+"' value='"+ row['name'] + "'>  ";
    result += "<span id='"+row['name']+"'";

    if ( row['type'] == Register_COIL || row['type'] == Register_HR) {
        result += " class='editable'";
    }
    result += "></span></td>";
    result += "</tr>";
    return result;
}

function addTableHeadForModbuses() {
    var headers = [];
    headers.push("<tr>");
    headers.push("<th>name</th>");
    headers.push("<th>desription</th>");
    headers.push("<th>type</th>");
    headers.push("<th>values</th>");
    headers.push("</tr>");
    $('#main_table_head').html(headers.join(""));
}

function getCurrentSubsystem(name) {
    setTitle(name);
    addTableHeadForSubsystems(subsystems[name]);
    clearAllCheckboxes();

    $.getJSON( urlCurrentSubsystem+name, function(data) {

        var items = [];
      
        $.each( data.variables, function() {
            if (isContentFilter( this['name'] + this['description'] ) && isShowSelected(this['name'])) {
                items.push( getRowVariable( this, name) );
            }
        });
        items.sort();
        $('#main_table_body').html( items.join(""));    
        
        currentSubsystem = name;
        startUpdatingSubsystem();
    }); 
}

function getCurrentModbus(name) {
    setTitle('[Modbus] ' + name);
    addTableHeadForModbuses();
    clearAllCheckboxes();
    currentModbus = name;
    $.getJSON( urlCurrentModbus+name, function(data) {
        var items = [];
        $.each( data.registers, function( ) {
            if (isContentFilter( this['name'] + this['desc'] ) && isShowSelected(this['name']) )  {
                items.push( getRowRegister(this) );
            }
        });
        items.sort();
        $('#main_table_body').html( items.join(""));     
        
        startUpdatingModbus();
    });
}

function updateValuesModbus() {
    $.getJSON( urlValuesModbus + currentModbus, function(data){
        for ( var j=0; j < data['values'].length; j++  ) {
            var variableId = "#"+data['values'][j]['name'];
            $(variableId).html(data['values'][j]['value']);
        }
    });
}

function updateValuesSubsystem() {
    for ( var i=0; i < subsystems[currentSubsystem].length; i++) {
        $.getJSON( urlValuesSubsystem+currentSubsystem+":"+subsystems[currentSubsystem][i], function(data) {
            const nameForID = data['name'].replace(/\./g, '');
            var variableId;
            for ( var j=0; j < data['values'].length; j++  ) {
                variableId = "#"+nameForID+data['values'][j]['name'];
                variableId=variableId.replace(':', '');
                $(variableId).html(data['values'][j]['value'][0]);
            }
        });
    }
}

function stopInterval() {
    if (updateIntervalId != "") {
        clearInterval(updateIntervalId);
        updateIntervalId = "";
    }
}

function stopUpdatingSubsystem() {
    currentSubsystem = "";
    stopInterval();
}

function stopUpdatingModbus() {
    currentModbus = "";
    stopInterval();    
}

function startUpdatingSubsystem() {
    if ( currentSubsystem != "") {
        stopUpdatingModbus();
        updateValuesSubsystem();
        updateIntervalId = setInterval( updateValuesSubsystem, timeInterval);
    }    
}

function startUpdatingModbus() {
    if (currentModbus != "") {
        stopUpdatingSubsystem();
        updateValuesModbus();
        updateIntervalId = setInterval( updateValuesModbus, timeInterval);
    }
}

function showSubsystems() {
    $.getJSON( urlSubs, function( data ) {
        var items = [];
        subsystems = [];

        $.each( data.routs, function() {
            subsystems[this['name']] = this['ips'];
            items.push( "<li class='nave-item'> <a class='nav-link' href='#' id='" + this['name'] + "'  >" + this['name'] + "</a></li>" );
        });

        items.sort();
       
        $( "#nav-left-subsystems").html(items.join( "" ));

        getCurrentSubsystem(defaultSubsystem);
      });      
}

function showModbuses() {
    $.getJSON( urlModbuses, function( data ) {
        var items = [];
        $.each( data.modbuses, function( ) {
            items.push( "<li class='nave-item'> <a class='nav-link' href='#' id='" + this['name'] + "'  >" + this['name'] + "</a></li>" );
        });
        items.sort();
        $( "#nav-left-modbuses").html(items.join( "" ));
      }); 
}

function updateContent() {
    if ( currentModbus.length > 0 ) {
        getCurrentModbus(currentModbus);     
    }
    else if ( currentSubsystem.length > 0) {
        getCurrentSubsystem(currentSubsystem);
    }
}

function setRemoteValue(spanId) {
    var id = '#'+spanId.replace(/&name=/i,'').replace(/\./gi,'').replace(/:/, '');
    var oldValue = $(id).html();
    var newValue = prompt("Enter new value: ", oldValue);

    if (newValue != null) {
        var url = ( currentModbus.length > 0 ) ? urlSetModbusValue: urlSetSubsystemValue;
        var data=( currentModbus.length > 0 ) ? 'modbus='+currentModbus+"&name="+spanId : 'subsystem='+spanId;
        data += "&value=" + newValue;
        $.ajax({
            url: url,
            data: data
        });
    }
}

$(document).ready(function(){
    showSubsystems();
    showModbuses();

    clearAllCheckboxes();
    $('#content-filter').val('');

    $("#nav-left-subsystems").on('click', '.nave-item a', function(){
        clearSelectedVariables();
        getCurrentSubsystem($(this).attr('id'));
    });

    $("#nav-left-modbuses").on('click', '.nave-item a', function(){
        clearSelectedVariables();
        getCurrentModbus($(this).attr('id'))
    });

    $(document).on('change','.checkboxes', function(){
        clickToCheckbox(this);
    });

    $("#content-filter").on("input", function() {
        contentFilter = this.value;
        updateContent();
    });

    $("#start-chart").click(function(){
        startChart();
    });

    $('#clear-checkboxes').click(function(){
        clearAllCheckboxes();
    });

    $(document).on('change','.'+classSelectVariable, function(){
        if ( this.checked ) {
            selectedItems.push(this.id);
        }
        else {
            for ( var i=0; i < selectedItems.length; i++) {
                if ( selectedItems[i] === this.id ) {
                    selectedItems.splice(i, 1);
                    break;
                }
            }
            if (selectedItems.length == 0) {
                $('#show-selected-variables').removeClass('active');
            }
            updateContent();   
        }
    });

    $('#clear-selected-variables').click(function() {
        clearSelectedVariables();
        updateContent();
    });

    $('#show-selected-variables').click(function() {
        if ( selectedItems.length > 0) {
            if (!isSelectedActive) {
                isSelectedActive = true;
                $(this).addClass('active');
            }
            else {
                isSelectedActive = false;
                $(this).removeClass('active');
            }   
            updateContent();     
        }
        else {
            alert('Не выбрано переменных для отображения!');
            updateContent();   
            $(this).removeClass('active');
        }
    });

    $(document).on('click', '.editable', function() {
//        setRemoteValue(this.id, this.innerHTML);        
        setRemoteValue(this.id);
    });
});
