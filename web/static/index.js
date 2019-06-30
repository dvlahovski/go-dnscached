$(document).ready(function() {
    $(".delete-button").click(function () {
        if(!confirm('Сигурни ли сте?')) {
            return;
        }
        $.ajax({
            url: "http://" + api_address + "/cache/delete",
            data: {
                key: $(this).attr("data-key"),
            },
            type: "GET",
            crossDomain: true,
            error: function (xhr, status) {
                alert("Грешка при изтриване!");
            }
        });
        location.reload();
    });

    $("#add-form").submit(function(event) {
        $.ajax({
            url: "http://" + api_address + "/cache/insert",
            data: {
                key: $("#add-url").val(),
                type: $("#add-type").val(),
                value: $("#add-ip").val(),
                ttl: $("#add-ttl").val(),
            },
            type: "GET",
            crossDomain: true,
            error: function (xhr, status) {
                alert("Грешка при добавяне!");
            }
        });
        // location.reload();
        // event.preventDefault();
    });
});