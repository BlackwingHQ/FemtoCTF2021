$(function () {
    $('#login_btn').click(function () {
        if ($('#typePassword').val() === "") {
            $('#out_form').val("Please enter a password...")
        } else {
            var data = {} 
            data["password"] = $('#typePassword').val()
            $('#typePassword').val("")
            var body = JSON.stringify(data)
            $.ajax({
                url: "https://spring-feather-9233.fly.dev/login",
                contentType: "application/json",
                data: body,
                dataType: "json",
                type: 'POST',
                success: function (response) {
                    data["token"] = response.token
                    var body = JSON.stringify(data)
                    $.ajax({
                        url: "https://spring-feather-9233.fly.dev/whoami",
                        contentType: "application/json",
                        data: body,
                        dataType: "json",
                        type: 'POST',
                        success: function (response) {
                            $('#out_form').val("Welcome " + response.data + "! \n(This is not the flag)")
                        },
                        error: function (response) {
                            $('#out_form').val("Bad token")
                        }
                        });
                },
                error: function (response) {
                    $('#out_form').val("Incorrect password.\n(psst. check out the code on GitHub)")
                }
            });
        }
    });
});