/**
 * Playwright.dev config.
 * 
 */
import { defineConfig } from '@playwright/test';

const testDir = 'tests/e2e-js';
const puppeteerOptions = process.env.CI ? {
  args: ['--no-sandbox', '--disable-setuid-sandbox']
} : {};

export default defineConfig({
  testDir,
  projects: [{
    name: 'fixtures',
    testMatch: /fixture\.test\.js/
  }, {
    name: 'fixtures-firefox',
    use: {
      browserName: 'firefox'
    },
    testMatch: /fixture\.test\.js/
  }, {
    name: 'fixtures-webkit',
    use: {
      browserName: 'webkit'
    },
    testMatch: /fixture\.test\.js/
  }, {
    name: 'api-chromium',
    use: {
      launchOptions: {
        ...puppeteerOptions
      }
    },
    testMatch: 'api/**/*.test.js',
    workers: 6,
    dependencies: ['fixtures']
  }, {
    name: 'api-firefox',
    use: {
      browserName: 'firefox'
    },
    testMatch: 'api/**/*.test.js',
    workers: 6,
    dependencies: ['fixtures-firefox']
  }, {
    name: 'api-webkit',
    use: {
      browserName: 'webkit'
    },
    testMatch: 'api/**/*.test.js',
    workers: 6,
    dependencies: ['fixtures-webkit']
  }]
});
