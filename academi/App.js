import React, { useEffect } from 'react';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { View, Text, StyleSheet } from 'react-native';

import ChatScreen from './screens/ChatScreen';
import CommunityScreen from './screens/CommunityScreen';
import DocsScreen from './screens/DocsScreen';
import GuideScreen from './screens/GuideScreen';
import ProfileScreen from './screens/ProfileScreen';
import useAppStore from './store/appStore';

const Tab = createBottomTabNavigator();

function TabBarIcon({ name, focused, color }) {
  const icons = {
    Chat: '💬',
    Community: '👥',
    Docs: '📄',
    Guide: '📚',
    Profile: '👤',
  };
  return (
    <View style={[styles.tabIcon, focused && styles.tabIconFocused]}>
      <Text style={[styles.iconText, focused && styles.iconTextFocused]}>
        {icons[name]}
      </Text>
    </View>
  );
}

function TabBarLabel({ focused, name }) {
  return (
    <Text style={[
      styles.tabLabel,
      focused && styles.tabLabelFocused
    ]}>
      {name}
    </Text>
  );
}

export default function App() {
  const { theme, initializeTheme } = useAppStore();

  useEffect(() => {
    initializeTheme();
  }, []);

  return (
    <SafeAreaProvider>
      <NavigationContainer>
        <Tab.Navigator
          screenOptions={({ route }) => ({
            tabBarIcon: ({ focused, color }) => (
              <TabBarIcon name={route.name} focused={focused} color={color} />
            ),
            tabBarLabel: ({ focused }) => (
              <TabBarLabel name={route.name} focused={focused} />
            ),
            headerShown: false,
            tabBarStyle: {
              backgroundColor: theme.bgPrimary,
              borderTopColor: theme.borderSubtle,
              height: 80,
              paddingBottom: 16,
              paddingTop: 8,
            },
            tabBarActiveTintColor: '#5B8CFF',
            tabBarInactiveTintColor: '#A8B2D1',
            tabBarItemStyle: {
              borderRadius: 20,
              justifyContent: 'center',
              alignItems: 'center',
            },
          })}
        >
          <Tab.Screen name="Chat" component={ChatScreen} />
          <Tab.Screen name="Community" component={CommunityScreen} />
          <Tab.Screen name="Docs" component={DocsScreen} />
          <Tab.Screen name="Guide" component={GuideScreen} />
          <Tab.Screen name="Profile" component={ProfileScreen} />
        </Tab.Navigator>
      </NavigationContainer>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  tabIcon: {
    height: 28,
    justifyContent: 'center',
    alignItems: 'center',
  },
  tabIconFocused: {
    transform: [{ scale: 1.1 }],
  },
  iconText: {
    fontSize: 20,
    fontWeight: '300',
    lineHeight: 24,
  },
  iconTextFocused: {
    fontWeight: '600',
  },
  tabLabel: {
    fontSize: 11,
    fontWeight: '500',
  },
  tabLabelFocused: {
    fontWeight: '700',
  },
});