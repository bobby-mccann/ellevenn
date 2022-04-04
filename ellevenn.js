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
    const elements = {};

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
        for (let el of elements[loc.context]) {
            el.innerText = loc.translated;
            el.insertBefore(editButton(loc.context), el.firstChild.nextSibling);
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

        let node;
        // For each text node on the page:
        while (node = walker.nextNode()) {
            let text = node.nodeValue;
            const match = text.match(/\|((\w+\.)+\w+)\[([^|]|)+]\|/);
            if (!match) {continue;}
            const context = match[0].substring(1, match[0].length-1);
            const args = match.slice(1);

            const el = document.createElement('span');
            el.innerText = context;

            node.nodeValue = "";

            const [before, after] = text.split(context);
            const bNode = document.createTextNode(before.replace('|', ''));
            const aNode = document.createTextNode(after.replace('|', ''));
            node.parentNode.insertBefore(aNode, node);
            node.parentNode.insertBefore(el, aNode);
            node.parentNode.insertBefore(bNode, el);

            registerElement(context, el);
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

    function registerElement(context, el) {
        if (elements[context] === undefined) {
            elements[context] = [];
        }
        elements[context].push(el);
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
        const newText = prompt(`Enter a new translation for ${context}:`);
        if (newText !== null && newText.length > 0) {
            // TODO: call the API
            localisations[context].translated = newText;
            setLocalisation(localisations[context]);
        }
    }
})();