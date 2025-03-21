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
    document.querySelector('#start-again-button').textContent = 'Request More Tokens';
    if (requestBody.get('paratime') === 'emerald') {
      document.querySelector('#add-emerald-button').style.display = 'block';
    }
    if (requestBody.get('paratime') === 'sapphire') {
      document.querySelector('#add-sapphire-button').style.display = 'block';
    }
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

// Add Sapphire to MetaMask
document.querySelector('#add-sapphire-button').addEventListener('click', () => {
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
document.querySelector('#add-emerald-button').addEventListener('click', () => {
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
