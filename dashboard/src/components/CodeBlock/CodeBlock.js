import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "@material-ui/core/styles";
import { emphasize } from "@material-ui/core/styles/colorManipulator";
import Typography from "@material-ui/core/Typography";
import CodeHighlight from "../CodeHighlight/CodeHighlight";

const styles = theme => ({
  root: {
    fontFamily: theme.typography.monospace.fontFamily,
    overflowX: "scroll",
    userSelect: "text",
    tabSize: 2,
  },
  background: {
    backgroundColor: emphasize(theme.palette.background.paper, 0.01875),
  },
  highlight: {
    color:
      theme.palette.type === "dark"
        ? theme.palette.secondary.light
        : theme.palette.secondary.dark,
    "& $background": {
      backgroundColor: emphasize(theme.palette.text.primary, 0.05),
    },
  },
  scaleFont: {
    // Browsers tend to render monospaced fonts a little larger than intended.
    // Attempt to scale accordingly.
    fontSize: "0.8125rem", // TODO: Scale given fontSize from theme?
  },
  wrap: { whiteSpace: "pre-wrap" },
});

class CodeBlock extends React.Component {
  static propTypes = {
    background: PropTypes.bool,
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    component: PropTypes.oneOfType([PropTypes.string, PropTypes.func]),
    children: PropTypes.node.isRequired,
    highlight: PropTypes.bool,
    scaleFont: PropTypes.bool,
  };

  static defaultProps = {
    background: false,
    component: "pre",
    className: "",
    highlight: false,
    scaleFont: true,
  };

  render() {
    const {
      background,
      classes,
      className: classNameProp,
      children,
      highlight,
      scaleFont,
      ...props
    } = this.props;

    const className = classnames(classes.root, classNameProp, {
      [classes.background]: background,
      [classes.scaleFont]: scaleFont,
      [classes.highlight]: highlight,
    });

    // TODO: make highlight be a prop used to specify the language
    // one of instead of a free string
    // on highlight component, for polling, need to make sure it will also update

    return (
      <Typography className={className} {...props}>
        <CodeHighlight language="properties" code={children}>
          {code => (
            <code
              className={classes.wrap}
              dangerouslySetInnerHTML={{ __html: code }}
            />
          )}
        </CodeHighlight>
      </Typography>
    );
  }
}

export default withStyles(styles)(CodeBlock);
