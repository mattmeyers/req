request {
    method = "POST"
    path = "/echo"
    headers = {
        Content-Type = "application/json"
    }
    body = <<-BODY
        {
            "foo": "bar"
        }
    BODY
}

response {
    assert "Status code" {
        expr = "res.code == 200"
    }
    assert "Content-Type header" {
        expr = "res.headers.Content-Type == application/json; charset=utf-8"
    }
    assert "Content-Length header" {
        expr = "res.headers.Content-Length > 0"
    }
    assert "Body" {
        expr = <<-BODY
            res.body == {
                "foo": "bar"
            }
        BODY
    }
}