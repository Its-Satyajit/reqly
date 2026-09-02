import { App } from "@reqly/frontend";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { initRequestBridge } from "./bridge";
import "../../../../frontend/src/index.css";

initRequestBridge();

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<App />
	</StrictMode>,
);
