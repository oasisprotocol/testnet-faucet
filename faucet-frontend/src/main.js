// @ts-check

function showResponseStatus(error, status, requestBody) {
  document.querySelector('#request-form').style.display = 'none';
  document.querySelector('#response-display').style.display = 'block';
  document.querySelector('#response-display-text').style.textAlign = 'center';
  if (error) {
    document.querySelector('#response-display-text').textContent = error;
    document.querySelector('#response-display-text').style.color = '#d1001f';
  } else {
    document.querySelector('#response-display-text').textContent = status;
    document.querySelector('#response-display-text').style.color = '#4BB543';
    if (requestBody.get('paratime') === 'emerald') {}
    if (requestBody.get('paratime') === 'sapphire') {}
  }
}
function showLoading(bool) {
  document.querySelector('#request-form-submit').disabled = bool;
}

/** Support /?paratime=sapphire */
function preselectParatimeFromURL() {
  const paratimeInUrl = new URL(window.location.href).searchParams.get('paratime');
  const paratimeSelect =
      /** @type {HTMLSelectElement} */
      (document.querySelector('#paratime'));

  for (const option of paratimeSelect.options) {
    if (option.value === paratimeInUrl) {
      option.selected = true;
    }
  }
}

document.querySelector('#request-form').addEventListener('submit', (event) => {
  try {
    showLoading(true);

    const form =
      /** @type {HTMLFormElement} */
      (event.currentTarget);
    const url = form.action;
    const requestBody = new URLSearchParams(new FormData(form))

    fetch(url, {
      method: 'POST',
      body: requestBody,
      headers: {
        Accept: 'application/json',
      },
    })
      .then(response => response.json())
      .then((responseJson) => {
        showLoading(false);
        showResponseStatus(null, responseJson.result, requestBody);
      }, (error) => {
        showLoading(false);
        showResponseStatus(error, null, requestBody);
      });

    // Only prevent native form POST if no errors were thrown until `fetch`
    event.preventDefault();
  } catch (error) {
    // Allow native form to POST instead.
    showLoading(false);
    console.error(error);
  }
});

preselectParatimeFromURL();
