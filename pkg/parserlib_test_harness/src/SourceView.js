import React from "react";
import { formatSpan } from './span';
import classNames from "classnames";
import "./GrammarView.css";
import "./SourceView.css";

// Render a syntax-highlighted view of the source from the trace.
// Highlight hovered spans.

export class SourceView extends React.Component {
  render() {
    return (
      <div className="source-view">
        <SourceViewNode {...this.props} />
      </div>
    );
  }
}

class SourceViewNode extends React.Component {
  render() {
    const {
      trace,
      grammar,
      highlightedSpan,
      onHighlightSpan,
      highlightedRuleID,
      onHighlightRule,
    } = this.props;

    const formattedSpan = trace ? formatSpan(trace) : null;
    const isHighlightedSpan = formattedSpan === highlightedSpan;

    function highlightWrapper(element) {
      return (
        <span
          className={classNames("source-span", {
            highlighted: isHighlightedSpan,
          })}
          onMouseOver={() => {
            onHighlightSpan(formattedSpan, true);
            onHighlightRule(trace.RuleID, true);
          }}
          onMouseOut={() => {
            onHighlightSpan(formattedSpan, false);
            onHighlightRule(trace.RuleID, false);
          }}
        >
          {element}
        </span>
      )
    }

    if (!trace) {
      return ""; // un-filled-in sequence items
    }

    const highlightProps = {
      onHighlightSpan: onHighlightSpan,
      highlightedSpan: highlightedSpan,
      onHighlightRule: onHighlightRule,
      highlightedRuleID: highlightedRuleID,
    };

    const rule = grammar.RulesByID[trace.RuleID];
    switch (rule.RuleType) {
      case "SEQUENCE":
        return (
          <span>
            {trace.ItemTraces.map((itemTrace, idx) => (
              <SourceViewNode
                key={idx}
                trace={itemTrace}
                grammar={grammar}
                {...highlightProps}
              />
            ))}
          </span>
        );
      case "CHOICE":
        return (
          <SourceViewNode
            trace={trace.ChoiceTrace}
            grammar={grammar}
            {...highlightProps}
          />
        );
      case "REF": {
        return (
          <SourceViewNode
            trace={trace.RefTrace}
            grammar={grammar}
            {...highlightProps}
          />
        );
      }
      case "KEYWORD":
        return highlightWrapper(
          <span className="rule-keyword">{rule.Keyword}</span>
        );
      case "REGEX":
        return highlightWrapper(
          <span
            className="rule-regex"
            style={{ whiteSpace: "pre" }}
          >
            {trace.RegexMatch}
          </span>
        );
      case "MAP":
        return (
          <SourceViewNode
            trace={trace.InnerTrace}
            grammar={grammar}
            {...highlightProps}
          />
        );
      case "SUCCEED":
        return null;
      default:
        console.error(trace);
        return <pre>{JSON.stringify(trace)}</pre>
    }
  }
}