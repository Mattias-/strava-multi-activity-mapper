import { render } from "preact";
import "preact/debug";
import { App } from "./app.tsx";
import "./index.css";
const root = document.getElementById("app");
if (root != null) {
  render(<App />, root);
}
