request {
  method = "GET"
  path = "${env.base_url}/ping"
}

response {
    assert "Status code" {
      expr = "res.code == 200"
    }
    assert "Conent-Type Header" {
      expr = "res.headers.Content-Type == text/plain"
    }
    assert "Content-Length Header" {
      expr = "res.headers.content-length > 0"
    }
    assert "Body" {
      expr = "res.body == pong"
    }
}