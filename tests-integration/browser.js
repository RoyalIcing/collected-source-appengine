const test = require("ava");
const puppeteer = require("puppeteer");

const baseURL = "http://localhost:8080";

exports.initBrowser = function() {
  test.before("create browser", async t => {
    const browser = await puppeteer.launch();
    t.context.browser = browser;

    t.context.onPage = async path => {
      const page = await browser.newPage();
      page.setDefaultNavigationTimeout(5000);

      await page.goto(baseURL + path);
      return page;
    };
  });

  test.after("close browser", async t => {
    await t.context.browser.close();
  });

  test.beforeEach(async t => {});
}
