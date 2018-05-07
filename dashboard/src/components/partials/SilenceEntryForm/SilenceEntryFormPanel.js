import React from "react";
import PropTypes from "prop-types";
import { withStyles } from "material-ui/styles";

import Grid from "material-ui/Grid";

import ExpansionPanel, {
  ExpansionPanelDetails,
  ExpansionPanelSummary,
} from "material-ui/ExpansionPanel";
import Typography from "material-ui/Typography";

import ExpandMoreIcon from "material-ui-icons/ExpandMore";

const StyledExpansionPanelSummary = withStyles(() => ({
  content: { maxWidth: "100%" },
}))(ExpansionPanelSummary);

const StyledExpansionPanelDetails = withStyles(() => ({
  root: { flexDirection: "column" },
}))(ExpansionPanelDetails);

class SilenceEntryFormPanel extends React.PureComponent {
  static propTypes = {
    title: PropTypes.node,
    summary: PropTypes.node,
    hasDefaultValue: PropTypes.bool,
    children: PropTypes.node,
  };

  static defaultProps = {
    title: undefined,
    summary: undefined,
    hasDefaultValue: false,
    children: undefined,
  };

  render() {
    const { title, summary, children, hasDefaultValue } = this.props;

    return (
      <ExpansionPanel defaultExpanded={hasDefaultValue}>
        <StyledExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
          <Grid container>
            <Grid item xs={4}>
              <Typography variant="body2" noWrap>
                {title}
              </Typography>
            </Grid>
            <Grid item xs={8}>
              <Typography
                color={hasDefaultValue ? "textSecondary" : "secondary"}
                noWrap
              >
                {summary}
              </Typography>
            </Grid>
          </Grid>
        </StyledExpansionPanelSummary>
        <StyledExpansionPanelDetails>{children}</StyledExpansionPanelDetails>
      </ExpansionPanel>
    );
  }
}

export default SilenceEntryFormPanel;
