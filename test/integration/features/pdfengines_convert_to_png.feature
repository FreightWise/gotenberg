Feature: /forms/pdfengines/convert-to-png

  Scenario: POST /forms/pdfengines/convert-to-png (Single PDF to PNG)
    Given I have a default Gotenberg container
    When I make a "POST" request to Gotenberg at the "/forms/pdfengines/convert-to-png" endpoint with the following form data and header(s):
      | files | testdata/Blank+pdf+1.pdf | file |
    Then the response status code should be 200
    Then the response header "Content-Type" should be "application/zip"
    Then there should be the following file(s) in the response:
      | page-1.png |
      | page-2.png |
      | page-3.png |

  Scenario: POST /forms/pdfengines/convert-to-png (Multiple PDFs to PNG)
    Given I have a default Gotenberg container
    When I make a "POST" request to Gotenberg at the "/forms/pdfengines/convert-to-png" endpoint with the following form data and header(s):
      | files | testdata/page_1.pdf      | file |
      | files | testdata/Blank+pdf+1.pdf | file |
    Then the response status code should be 200
    Then the response header "Content-Type" should be "application/zip"
    Then there should be the following file(s) in the response:
      | page-1.png |
      | page-2.png |
      | page-3.png |
      | page-4.png |

  Scenario: POST /forms/pdfengines/convert-to-png (Bad Request - No Files)
    Given I have a default Gotenberg container
    When I make a "POST" request to Gotenberg at the "/forms/pdfengines/convert-to-png" endpoint with the following form data and header(s):
      | dummy | value | field |
    Then the response status code should be 400
    Then the response header "Content-Type" should be "text/plain; charset=UTF-8"
    Then the response body should match string:
      """
      Invalid form data: no form file found for extensions: [.pdf]
      """

  Scenario: POST /forms/pdfengines/convert-to-png (Output Filename)
    Given I have a default Gotenberg container
    When I make a "POST" request to Gotenberg at the "/forms/pdfengines/convert-to-png" endpoint with the following form data and header(s):
      | files                     | testdata/Blank+pdf+1.pdf | file   |
      | Gotenberg-Output-Filename | converted-images         | header |
    Then the response status code should be 200
    Then the response header "Content-Type" should be "application/zip"
    Then there should be the following file(s) in the response:
      | converted-images.zip |
      | page-1.png          |
      | page-2.png          |
      | page-3.png          |

  Scenario: POST /forms/pdfengines/convert-to-png (Gotenberg Trace)
    Given I have a default Gotenberg container
    When I make a "POST" request to Gotenberg at the "/forms/pdfengines/convert-to-png" endpoint with the following form data and header(s):
      | files           | testdata/Blank+pdf+1.pdf           | file   |
      | Gotenberg-Trace | forms_pdfengines_convert_to_png    | header |
    Then the response status code should be 200
    Then the response header "Content-Type" should be "application/zip"
    Then the response header "Gotenberg-Trace" should be "forms_pdfengines_convert_to_png"
    Then the Gotenberg container should log the following entries:
      | "trace":"forms_pdfengines_convert_to_png" |
