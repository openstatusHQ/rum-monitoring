import { zValidator } from "@hono/zod-validator";
import { browserName, detectOS } from "detect-browser";
import { Hono } from "hono";
import { z } from "zod";

type Bindings = {
  RUM_SERVER_URL: string;
};

const app = new Hono<{ Bindings: Bindings }>();

const schema = z.object({
  dsn: z.string(),
  event_name: z.string(),
  href: z.string(),
  id: z.string(),
  language: z.string(),
  os: z.string(),
  page: z.string(),
  screen: z.string(),
  speed: z.string(),
  value: z.number(),
});

const chSchema = schema.extend({
  browser: z.string(),
  city: z.string(),
  country: z.string(),
  continent: z.string(),
  device: z.string(),
  region_code: z.string(),
  timezone: z.string(),
});

app.get("/", (c) => {
  return c.text("Hello Hono!");
});

app.post("/ingest", zValidator("json", schema), async (c) => {
  const data = c.req.valid("json");
  const userAgent = c.req.header("user-agent") || "";

  const country = c.req.header("cf-ipcountry") || "";
  const city = c.req.header("cf-ipcity") || "";
  const region_code = c.req.header("cf-region-code") || "";
  const timezone = c.req.header("cf-timezone") || "";
  const browser = browserName(userAgent);

  const os = detectOS(userAgent);
  const payload = chSchema.parse({
    ...data,
    country,
    city,
    timezone,
    region_code,
    os,
  });

  const insert = async () => {
    console.log(payload);
    await fetch("http://localhost:8080/ingest", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });
  };
  c.executionCtx.waitUntil(insert());
  return c.json({ status: "ok" }, 200);
});

export default app;
