// @ts-check

function showResponseStatus(status) {
  document.querySelector('#request-form').style.display = 'none';
  document.querySelector('#response-display').style.display = 'block';
  document.querySelector('#response-display-text').textContent = status;
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

    fetch(url, {
      method: 'POST',
      body: new URLSearchParams(new FormData(form)),
      headers: {
        Accept: 'application/json',
      },
    })
      .then(response => response.json())
      .then((responseJson) => {
        showLoading(false);
        showResponseStatus(responseJson.result);
      }, (error) => {
        showLoading(false);
        showResponseStatus(error);
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
