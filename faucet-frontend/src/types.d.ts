interface ParentNode {
  // Non nullable
  querySelector<K extends keyof HTMLElementTagNameMap>(selectors: K): HTMLElementTagNameMap[K]
  querySelector<K extends keyof SVGElementTagNameMap>(selectors: K): SVGElementTagNameMap[K]
  querySelector<K extends keyof MathMLElementTagNameMap>(selectors: K): MathMLElementTagNameMap[K]
  querySelector<E extends Element = HTMLElement>(selectors: string): E
}

interface Window {
  ethereum?: {
    request(request: {
        method: string;
        params?: Array<any> | Record<string, any>;
    }): Promise<any>;
  }
}

declare const process: {
  env: {
    DOCUMENT_TITLE: string,
    DOCUMENT_DESCRIPTION: string,
    CARD_TITLE: string,
    DROP_LIMIT_TEXT: string,
    REQUEST_AMOUNT: string,

    FAUCET_BALANCE_API: string,
    FAUCET_BALANCE_EXPLORER: string,

    CAPTCHA_SITE_KEY: string,
  }
}
