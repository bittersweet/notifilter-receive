function duplicateSettings() {
    var removeRule = '<input type="button" name="removeRule" value="Remove rule" class="btn">';
    el = $(".ruleSet:first").clone();
    el.append(removeRule);
    el.appendTo("#rulesContainer");
}

function generateSettingsFromFields() {
    rules = []
    $.each($(".ruleSet"), function(index, ruleSet) {
        var key = $(ruleSet).children("select[name=key]").val();
        var type = $(ruleSet).children("select[name=type]").val();
        var setting = $(ruleSet).children("select[name=setting]").val();
        var value = $(ruleSet).children("input[name=value]").val();

        config = {
            "key": key,
            "type": type,
            "setting": setting,
            "value": value,
        }

        rules.push(config);
    });

    return rules;
}

$(document).ready(function() {
    $("#notifierForm input[name=generateRules]").on("click", function(event) {
        event.preventDefault();

        rules = generateSettingsFromFields();
        input = JSON.stringify(rules, null, 2);

        $("#notifierForm #rules").text(input);
    })

    $("#notifierForm input[name=addRule]").on("click", function(event) {
        event.preventDefault();
        duplicateSettings();
    });

    $("#notifierForm").on("click",  "input[name=removeRule]", function(event) {
        event.preventDefault();
        $(this).parent().remove();
    });

    $("#notifierForm input[name=preview]").on("click", function(event) {
        event.preventDefault();

        var notifierClass = $("#notifierForm select[name=class]").val();
        var template = $("#template").val();
        payload = {
            class: notifierClass,
            template: template
        }

        $.post("/preview", payload, function(data) {
            console.log("output: ", data);
            $("#templatePreview").val(data);
        }, "text");
    });
});

