// ==UserScript==
// @name         ellevenn - In Browser Localisation
// @namespace    spareroom
// @version      0.1
// @description
// @author       Bobby McCann
// @match        https://local.spareroom.co.uk/*
// @icon         https://assets.spareroom.co.uk/favicon.png?v=1
// @grant        none
// ==/UserScript==

(function() {
    'use strict';
    const localisations = {};
    const listeners = {};

    doReplace(document.body);

    new MutationObserver(mutations => {
        for (const m of mutations) {
            for (const node of m.addedNodes) {
                doReplace(node);
            }
        }
    }).observe(document.body, {childList: true, subtree: true});

    function setLocalisation(loc) {
        localisations[loc.context] = loc;
        for (let l of listeners[loc.context]) {
            l(loc);
        }
    }

    function doReplace(fromNode) {
        const walker = document.createTreeWalker(
            fromNode,
            NodeFilter.SHOW_TEXT,
            {
                acceptNode: node => node.nodeValue.trim() ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_SKIP
            },
        );

        const regex = /\|((?:\w+\.)+\w+) \[(?:([^|]+)\|)*([^|]+)?]\|/;
        let node;
        // For each text node on the page:
        while (node = walker.nextNode()) {
            let text = node.nodeValue;
            const match = text.match(regex);
            if (!match) {continue;}
            const context = match[1];
            const args = match.slice(2).filter(v => v !== undefined);

            const orig = `|${context} [${args.join('|')}]|`;
            const [before, after] = text.split(orig);

            node.nodeValue = "";

            const el = document.createElement('span');
            el.innerText = context;

            const bNode = document.createTextNode(before);
            const aNode = document.createTextNode(after);
            node.parentNode.insertBefore(aNode, node);
            node.parentNode.insertBefore(el, aNode);
            node.parentNode.insertBefore(bNode, el);

            registerElement(context, args, el);
            fetch(`http://localhost:8111/localisation/${context}`)
            .then((response) => {
                response.json().then(l => {
                    setLocalisation(l);
                })
                .catch(console.log);
            })
            .catch(console.log);
        }
    }

    function registerElement(context, args, el) {
        registerListener(context, loc => {
            const hasTranslation = loc.translated && loc.translated.length > 0;
            let newText = hasTranslation? loc.translated : loc.original;
            for (let arg of args) {
                newText = newText.replace("%s", arg);
            }
            // TODO: handle too many / too few %s

            el.innerText = newText;
            if (!hasTranslation) {
                el.style.backgroundColor = "RGBA(255, 255, 0, 0.4)";
            } else {
                el.style.backgroundColor = "";
            }
            el.insertBefore(editButton(loc.context), el.firstChild.nextSibling);
        });
    }

    function registerListener(context, onUpdate) {
        if (listeners[context] === undefined) {
            listeners[context] = [];
        }
        listeners[context].push(onUpdate);
    }

    function editButton(context) {
        const el = document.createElement('span');
        el.innerText = "📝";
        el.onclick = ev => {
            editLocalisation(context);
            ev.preventDefault();
        }
        return el;
    }

    function editLocalisation(context) {
        const newText = prompt(`Enter Translation:\nEN: ${localisations[context].original}`);
        if (newText !== null && newText.length > 0) {
            const newL = {
                context: localisations[context].context,
                original: localisations[context].original,
                translated: newText
            };

            fetch(`http://localhost:8111/localisation`,
                {
                    method: "POST",
                    body: JSON.stringify(newL)
                }
            )
            .then((response) => {
                response.json().then(l => {
                    setLocalisation(l);
                })
                .catch(console.log);
            })
            .catch(console.log);
        }
    }
})();