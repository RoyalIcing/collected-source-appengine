import test from "ava";
import { initBrowser } from "../browser";

initBrowser();

test("WHEN creating a post using form", async t => {
  const page = await t.context.onPage("/org:RoyalIcing/channel:design/posts");

  const newPostForm = await page.$("form[data-target='posts.createForm'");
  t.truthy(newPostForm, "Has create post form");
  const textarea = await newPostForm.$("textarea");
  await textarea.type("Post content");
  await Promise.all([
    // page.waitForNavigation(),
    newPostForm.click("button[value='submitPost']")
  ]);
});

test("WHEN viewing posts for a channel", async t => {
  const page = await t.context.onPage("/org:RoyalIcing/channel:design/posts");

  const heading = await page.$eval("h1 a", async ({ textContent }) => ({
    textContent
  }));
  t.deepEqual(
    await heading.textContent,
    "💬 design",
    "THEN the slug is the heading"
  );
});

test.skip("WHEN querying using developer section", async t => {
  const page = await t.context.onPage("/org:RoyalIcing/channel:design/posts");

  const developerSection = await page.$("[data-controller='developer']");
  t.truthy(developerSection, "THEN has developer section on the page");

  await developerSection.click("> details > summary");
  await developerSection.click("button");

  const queryResultCode = developerSection.$("code[data-target='developer.result']");
  t.truthy(queryResultCode, "THEN has query result on the page");

  const resultText = await page.waitForFunction(
    queryResultCode => {
      return queryResultCode.textContent;
    },
    { timeout: 3000 },
    queryResultCode
  );

  t.regex(await resultText.jsonValue(), /"data"/, "THEN has result data");
});
