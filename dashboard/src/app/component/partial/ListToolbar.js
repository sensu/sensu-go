import React from "react";
import PropTypes from "prop-types";

import WithWidth, { isWidthUp } from "/lib/component/util/WithWidth";

import Toolbar from "/app/component/partial/Toolbar";

const collapseAtWidth = "xs";

class ListToolbar extends React.PureComponent {
  static propTypes = {
    search: PropTypes.node.isRequired,
    toolbarItems: PropTypes.func.isRequired,
    widthProvider: PropTypes.node,
  };

  static defaultProps = {
    widthProvider: <WithWidth />,
  };

  renderSearch = ({ width: targetWidth }) => {
    const { search } = this.props;
    const { style: inlineStyleProp } = search.props;

    const width = isWidthUp(targetWidth, collapseAtWidth) ? "100%" : "50%";
    const style = { width, ...inlineStyleProp };
    return React.cloneElement(search, { style });
  };

  renderToolbarItems = ({ width }) => {
    const collapsed = width === collapseAtWidth;
    return this.props.toolbarItems({ width, collapsed });
  };

  render() {
    const { search, toolbarItems, widthProvider, ...props } = this.props;

    return (
      <widthProvider.type {...widthProvider.props}>
        {providerProps => (
          <Toolbar
            left={this.renderSearch(providerProps)}
            right={this.renderToolbarItems(providerProps)}
            {...props}
          />
        )}
      </widthProvider.type>
    );
  }
}

export default ListToolbar;
