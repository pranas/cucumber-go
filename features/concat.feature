Feature: String concat
  Scenario: foobar
    When you concat "foo" and "bar"
    Then you should have "foobar"

  Scenario: hello world
    When you concat "hello " and "world"
    Then you should have "hello world"
