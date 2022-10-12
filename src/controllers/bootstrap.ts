import { router, mount } from "@codewithkyle/router";

router.add("/*", "missing-page");

mount(document.body);
