// @ts-check

function showResponseStatus(status) {
  document.querySelector('#request-form').style.display = 'none';
  document.querySelector('#response-display').style.display = 'block';
  document.querySelector('#response-display-text').textContent = status;
}

document.querySelector('#request-form').addEventListener('submit', (event) => {
  try {
    document.querySelector('#request-form-submit').disabled = true;

    const form =
      /** @type {HTMLFormElement} */
      (event.currentTarget);
    const url = form.action;

    fetch(url, {
      method: 'POST',
      body: new URLSearchParams(new FormData(form)),
      headers: {
        Accept: 'application/json',
      },
    })
      .then(response => response.json())
      .then((responseJson) => {
        document.querySelector('#request-form-submit').disabled = false;
        showResponseStatus(responseJson.result);
      }, (error) => {
        document.querySelector('#request-form-submit').disabled = false;
        showResponseStatus(error);
      });

    // Only prevent native form POST if no errors were thrown until `fetch`
    event.preventDefault();
  } catch (error) {
    // Allow native form to POST instead.
    document.querySelector('#request-form-submit').disabled = false;
    console.error(error);
  }
});

