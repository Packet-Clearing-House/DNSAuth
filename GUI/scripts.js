
API_URL = "http://216.21.3.3:8080"

json = [["123.99.248.35/24","Customer1","true","true"],["31.178.129.217/24","Customer1","true","true”],[“1501:0121:0800:0000::/48","Customer1","true","true”],[“1640:0111:0561:0200::/48","Customer1","true","true"],["132.50.192.61","Customer2","false","false”],[“1056:340:14:6400:ad::1","Customer2","false","false"],["70.251.169.185","Customer6","false","false"],["124.221.36.60","Customer3","false","false"],["127.83.153.5","Customer5","false","false”],[“1241:400:14:6566:aa::1”,”Customer5","false","false"],["28.18.106.18","Customer4","false","false”],[“241.14.190.185”,”Customer1","false","false"]]

$(document).ready(function() {
    // $('#example').DataTable();
    var table = $('#example').DataTable( {
        dom: 'Bfrtip',
        scrollY:        '50vh',
        scrollCollapse: true,
        paging:         false,
        "ajax": API_URL + "/customers/list",
        "columns": [
            { "data": null },
            { "data": "name" },
            { "data": "ip" },
            { 
                "data": "prefix",
                render: function ( data, type, row ) {
                    if ( type === 'display' ) {
                        return '<input type="checkbox" class="prefix">';
                    }
                    return data;
                },
                className: "dt-body-center"                
            },
            { 
                "data": "asn",
                render: function ( data, type, row ) {
                    if ( type === 'display' ) {
                        return '<input type="checkbox" class="asn">';
                    }
                    return data;
                },
                className: "dt-body-center"                
            },
        ],
        rowCallback: function ( row, data ) {
            // Set the checked state of the checkbox in the table
            $('input.prefix', row).prop( 'checked', data.prefix == "true");
            $('input.asn', row).prop( 'checked', data.asn == "true");
        },
        "columnDefs": [ {
            "targets": 0,
            "data": null,
            sortable: false,
            "defaultContent": "",
            className: 'select-checkbox',
        } ],
        select: {
            style:    'multi',
            selector: 'td:first-child'
        },
        order: [[ 1, 'asc' ]],
        buttons: [
            {
                text: 'Delete',
                action: function ( e, dt, node, config ) {
                    var rows_selected = table.rows('.selected').data();
                    console.log(rows_selected);
                    rows_selected.each(function(v, i) {
                        $.get(API_URL + "/customers/delete?" + jQuery.param({ip:v.ip})).fail(function() {
                            alert("Something went wrong while deleting a customer...")
                        }).done(function() {
                            table.rows('.selected').remove().draw();
                        });
                    })
                }
            },
            // { extend: "create", editor: editor },
        ]
        // "dataSrc": function ( json ) {
        //     for ( var i=0, ien=json.length ; i<ien ; i++ ) {
        //       json[i][0] = '<a href="/message/'+json[i][0]+'>View message</a>';
        //     }
        //     console.log(json)
        //     return json;
        // },
        
        // // "ajax": "array.json",
        // // "dataSrc":  "",
        // // "data": json,
        // "columnDefs": [ {
        //     'targets': 0,
        //     'checkboxes': {
        //        'selectRow': true
        //     },
        //     // data: "",
        // }],
            // targets: 1,
        //     data: 1
        // },
        // {
        //     targets: 2,
        //     data: 0
        // },
        // {
        //     targets: 3,
        //     data: 2
        // },
        // {
        //     targets: 4,
        //     data: 3
        // } ],
        // "select": {
        //     'style': 'multi'
        //     // selector: 'td:first-child'
        // },
        // order: [[ 1, 'asc' ]]
    });    

    $('#example').on( 'change', 'input.prefix', function (e, t) {
        data = table.row($(this).closest('tr')).data()
        data.prefix = $(this).prop('checked') ? "true": "false"
        jQuery.param(data)
        // console.log($(this).row())
        $.get(API_URL + "/customers/update?" + jQuery.param(data)).fail(function() {
            alert("Something went wrong while updating a customer...")
        }).done(function() {
            table.rows('.selected').remove().draw();
        });
    } );

} );
