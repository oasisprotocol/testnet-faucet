// @ts-check

function $() {
  return {
    response_display: document.querySelector('#response_display'),
    response_display_text: document.querySelector('#response_display_text'),
    start_again_button: /** @type {HTMLButtonElement} */ (document.querySelector('#start_again_button')),
    add_emerald_button: /** @type {HTMLButtonElement} */ (document.querySelector('#add_emerald_button')),
    add_sapphire_button: /** @type {HTMLButtonElement} */ (document.querySelector('#add_sapphire_button')),
    request_form: document.querySelector('#request_form'),
    request_form_submit: /** @type {HTMLButtonElement} */ (document.querySelector('#request_form_submit')),
    paratime: /** @type {HTMLSelectElement} */ (document.querySelector('#paratime')),
    is_drained: /** @type {HTMLButtonElement} */ (document.querySelector('#is_drained')),
    is_almost_drained: /** @type {HTMLButtonElement} */ (document.querySelector('#is_almost_drained')),
  }
}


/**
 * @param {null | string} error
 * @param {null | string} success
 * @param {URLSearchParams} requestBody
 */
function showResponseStatus(error, success, requestBody) {
  $().request_form.style.display = 'none';
  $().response_display.style.display = 'block';
  $().response_display_text.style.textAlign = 'center';
  if (error) {
    $().response_display_text.textContent = error;
    $().response_display_text.style.color = '#d1001f';
  } else {
    $().response_display_text.textContent = success;
    $().response_display_text.style.color = '#4BB543';
    $().start_again_button.textContent = 'Request More Tokens';
    if (requestBody.get('paratime') === 'emerald') {
      $().add_emerald_button.style.display = 'block';
    }
    if (requestBody.get('paratime') === 'sapphire') {
      $().add_sapphire_button.style.display = 'block';
    }
  }
}
/** @param {boolean} bool */
function showLoading(bool) {
  $().request_form_submit.disabled = bool;
}

/** Support /?paratime=sapphire */
function preselectParatimeFromURL() {
  const paratimeInUrl = new URL(window.location.href).searchParams.get('paratime');
  for (const option of $().paratime.options) {
    if (option.value === paratimeInUrl) {
      option.selected = true;
    }
  }
}

$().request_form.addEventListener('submit', (event) => {
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

// Add Sapphire to MetaMask
$().add_sapphire_button.addEventListener('click', () => {
  if (!window.ethereum?.request) {
    return alert(
      'Have you installed MetaMask yet? If not, please do so.\n\nComputer: Once it is installed, you will be able to add the Sapphire ParaTime to your MetaMask.\n\nPhone: Open the website through your MetaMask Browser to add the Sapphire ParaTime.'
    );
  }

  const startTime = Date.now();
  window.ethereum
    .request({
      method: 'wallet_addEthereumChain',
      params: [
        {
          chainId: '0x5aff',
          chainName: 'Oasis Sapphire Testnet',
          nativeCurrency: {
            name: 'TEST',
            symbol: 'TEST',
            decimals: 18,
          },
          rpcUrls: [
            'https://testnet.sapphire.oasis.io',
            'wss://testnet.sapphire.oasis.io/ws',
          ],
          blockExplorerUrls: ['https://explorer.oasis.io/testnet/sapphire'],
        },
      ],
    })
    .then((response) => {
      const isAutomatedResponse = Date.now() - startTime < 100;
      if (response === null && isAutomatedResponse)
        alert('The Oasis Sapphire Testnet RPC already added.');
    });
});

// Add Emerald to MetaMask
$().add_emerald_button.addEventListener('click', () => {
  if (!window.ethereum?.request) {
    return alert(
      'Have you installed MetaMask yet? If not, please do so.\n\nComputer: Once it is installed, you will be able to add the Emerald ParaTime to your MetaMask.\n\nPhone: Open the website through your MetaMask Browser to add the Emerald ParaTime.'
    );
  }

  const startTime = Date.now();
  window.ethereum
    .request({
      method: 'wallet_addEthereumChain',
      params: [
        {
          chainId: '0xa515',
          chainName: 'Oasis Emerald Testnet',
          nativeCurrency: {
            name: 'TEST',
            symbol: 'TEST',
            decimals: 18,
          },
          rpcUrls: [
            'https://testnet.emerald.oasis.io',
            'wss://testnet.emerald.oasis.io/ws',
          ],
          blockExplorerUrls: ['https://explorer.oasis.io/testnet/emerald'],
        },
      ],
    })
    .then((response) => {
      const isAutomatedResponse = Date.now() - startTime < 100;
      if (response === null && isAutomatedResponse)
        alert('The Oasis Emerald Testnet RPC already added.');
    });
});

async function checkFaucetBalance() {
  if (!process.env.FAUCET_BALANCE_API) return
  const acc = await (await fetch(process.env.FAUCET_BALANCE_API)).json()
  if (BigInt(acc.available) < BigInt(process.env.REQUEST_AMOUNT) * BigInt('1000000000')) {
    $().is_drained.style.display = 'block'
  } else if (BigInt(acc.available) < BigInt(process.env.REQUEST_AMOUNT) * BigInt('1000000000') * BigInt('20')) {
    $().is_almost_drained.style.display = 'block'
  }
}

checkFaucetBalance()
