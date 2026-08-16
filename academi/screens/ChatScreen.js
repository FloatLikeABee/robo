import React, { useState, useEffect, useRef } from 'react';
import {
  View, Text, StyleSheet, FlatList, TextInput,
  SafeAreaView, StatusBar, TouchableOpacity, Dimensions,
  Animated, ScrollView, Image
} from 'react-native';
import ChatBubble from '../components/ChatBubble';
import Button from '../components/Button';
import NeuralGlow from '../components/NeuralGlow';
import AIResponseBlock from '../components/AIResponseBlock';
import TagChip from '../components/TagChip';
import useAppStore from '../store/appStore';

const { width, height } = Dimensions.get('window');

const QUICK_ACTIONS = [
  { id: 1, label: '⟐ Summarize', query: 'Summarize this topic' },
  { id: 2, label: '⍰ Explain', query: 'Explain in detail' },
  { id: 3, label: '⟑ Study Plan', query: 'Create a study plan' },
  { id: 4, label: '⏣ Examples', query: 'Give me examples' },
];

const SUGGESTED_TOPICS = [
  '#math', '#physics', '#cs', '#biology', '#chemistry', '#quantum'
];

export default function ChatScreen() {
  const { theme } = useAppStore();
  const [messages, setMessages] = useState([
    {
      id: 1,
      text: "Hello! I'm Academi, your AI study assistant. Ask me anything or pick a topic below!",
      isUser: false,
      hasSource: false,
    },
  ]);
  const [inputText, setInputText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [expandedMessageId, setExpandedMessageId] = useState(null);
  const [showSuggestions, setShowSuggestions] = useState(true);
  const flatListRef = useRef(null);
  const typingAnim = useRef(new Animated.Value(0)).current;
  const inputAnim = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    if (isTyping) {
      Animated.loop(
        Animated.sequence([
          Animated.timing(typingAnim, {
            toValue: 1,
            duration: 600,
            useNativeDriver: true,
          }),
          Animated.timing(typingAnim, {
            toValue: 0,
            duration: 600,
            useNativeDriver: true,
          }),
        ])
      ).start();
    }
  }, [isTyping]);

  const sendMessage = (text) => {
    if (text.trim() === '') return;

    setShowSuggestions(false);
    const userMessage = {
      id: Date.now(),
      text: text,
      isUser: true,
      hasSource: false,
    };

    setMessages(prev => [...prev, userMessage]);
    setInputText('');
    setIsTyping(true);

    setTimeout(() => {
      const aiResponse = {
        id: Date.now() + 1,
        text: generateAIResponse(text),
        isUser: false,
        hasSource: true,
        sources: [
          { title: 'Wikipedia', type: 'web' },
          { title: 'Research Paper', type: 'doc' },
        ],
      };

      setIsTyping(false);
      setMessages(prev => [...prev, aiResponse]);
    }, 1500 + Math.random() * 1000);
  };

  const generateAIResponse = (query) => {
    const responses = {
      summarize: "Here's a concise summary of the topic:\n\n• Key concept 1: The fundamental principles\n• Key concept 2: The practical applications\n• Key concept 3: The future implications\n\nWould you like me to elaborate on any specific point?",
      explain: "Let me break this down step by step:\n\n1. First, we need to understand the basic framework\n2. Then we look at how the components interact\n3. Finally, we apply this to real-world scenarios\n\nThis approach will help you build a solid understanding.",
      default: "Great question! Based on my analysis across multiple sources:\n\nThe core concept revolves around three main pillars:\n\n📌 Theory — The foundational principles\n📌 Practice — Real-world applications\n📌 Innovation — Emerging developments\n\nI've pulled information from academic papers, community discussions, and web sources. Tap on the sources below for full references.",
    };

    const lowerQuery = query.toLowerCase();
    if (lowerQuery.includes('summarize')) return responses.summarize;
    if (lowerQuery.includes('explain')) return responses.explain;
    return responses.default;
  };

  const handleQuickAction = (action) => {
    sendMessage(action.query);
  };

  const handleTopicPress = (topic) => {
    setInputText(`Tell me about ${topic}`);
  };

  const renderMessage = ({ item, index }) => {
    if (item.isUser) {
      return (
        <ChatBubble
          text={item.text}
          isUser={true}
          hasSource={false}
        />
      );
    }

    return (
      <AIResponseBlock
        text={item.text}
        sources={item.sources || []}
        expanded={expandedMessageId === item.id}
        onToggleExpand={() => setExpandedMessageId(
          expandedMessageId === item.id ? null : item.id
        )}
        onSave={() => console.log('Save:', item.id)}
        onShare={() => console.log('Share:', item.id)}
        onCopy={() => console.log('Copy:', item.id)}
        showActions={true}
      />
    );
  };

  const renderTypingIndicator = () => {
    if (!isTyping) return null;

    return (
      <View style={themedStyles.typingContainer}>
        <NeuralGlow isActive={isTyping} size={4} />
        <Animated.View style={{ opacity: typingAnim }}>
          <Text style={themedStyles.typingText}>Academi is thinking...</Text>
        </Animated.View>
      </View>
    );
  };

  const renderSuggestions = () => {
    if (!showSuggestions) return null;

    return (
      <View style={themedStyles.suggestionsContainer}>
        <Text style={themedStyles.suggestionsTitle}>Quick Actions</Text>
        <ScrollView horizontal showsHorizontalScrollIndicator={false} style={themedStyles.quickActionsRow}>
          {QUICK_ACTIONS.map((action) => (
            <TouchableOpacity
              key={action.id}
              style={themedStyles.quickActionButton}
              onPress={() => handleQuickAction(action)}
              activeOpacity={0.7}
            >
              <Text style={themedStyles.quickActionText}>{action.label}</Text>
            </TouchableOpacity>
          ))}
        </ScrollView>

        <Text style={themedStyles.suggestionsTitle}>Suggested Topics</Text>
        <View style={themedStyles.topicsRow}>
          {SUGGESTED_TOPICS.map((topic, index) => (
            <TouchableOpacity key={index} onPress={() => handleTopicPress(topic)} activeOpacity={0.6}>
              <TagChip label={topic} variant="glow" />
            </TouchableOpacity>
          ))}
        </View>
      </View>
    );
  };

  const themedStyles = StyleSheet.create({
    container: {
      flex: 1,
      backgroundColor: theme.colors.bgPrimary,
    },
    header: {
      flexDirection: 'row',
      justifyContent: 'space-between',
      alignItems: 'center',
      paddingHorizontal: 20,
      paddingVertical: 12,
      borderBottomWidth: 1,
      borderBottomColor: theme.colors.borderSubtle,
    },
    headerTitle: {
      color: theme.colors.textPrimary,
      fontSize: 20,
      fontWeight: '700',
      letterSpacing: 0.5,
    },
    headerRight: {
      flexDirection: 'row',
      alignItems: 'center',
      gap: 8,
    },
    headerStatus: {
      color: theme.colors.textSecondary,
      fontSize: 12,
    },
    suggestionsToggle: {
      paddingHorizontal: 12,
      paddingVertical: 6,
      borderRadius: 16,
      backgroundColor: theme.colors.surfaceGlass,
      borderWidth: 1,
      borderColor: theme.colors.borderSubtle,
    },
    suggestionsToggleText: {
      color: theme.colors.textPrimary,
      fontSize: 12,
      fontWeight: '600',
    },
    messagesContainer: {
      padding: 16,
      paddingBottom: 16,
      flexGrow: 1,
    },
    flatList: {
      flex: 1,
    },
    chatContent: {
      flex: 1,
    },
    suggestionsContainer: {
      marginBottom: 16,
    },
    suggestionsTitle: {
      color: theme.colors.textSecondary,
      fontSize: 13,
      fontWeight: '600',
      marginBottom: 10,
      textTransform: 'uppercase',
      letterSpacing: 1,
    },
    quickActionsRow: {
      marginBottom: 16,
    },
    quickActionButton: {
      paddingHorizontal: 16,
      paddingVertical: 10,
      borderRadius: 20,
      backgroundColor: theme.colors.surfaceGlass,
      marginRight: 8,
      borderWidth: 1,
      borderColor: theme.colors.borderSubtle,
    },
    quickActionText: {
      color: theme.colors.textPrimary,
      fontSize: 13,
      fontWeight: '500',
    },
    topicsRow: {
      flexDirection: 'row',
      flexWrap: 'wrap',
    },
    typingContainer: {
      flexDirection: 'row',
      alignItems: 'center',
      paddingHorizontal: 20,
      paddingVertical: 8,
    },
    typingText: {
      color: theme.colors.primary,
      fontSize: 12,
      marginLeft: 8,
      fontWeight: '500',
    },
    inputContainer: {
      paddingHorizontal: 12,
      paddingBottom: 12,
    },
    glassInputWrapper: {
      flexDirection: 'row',
      alignItems: 'flex-end',
      backgroundColor: theme.colors.surfaceGlass,
      borderRadius: 24,
      paddingHorizontal: 16,
      paddingVertical: 8,
      borderWidth: 1,
      borderColor: theme.colors.borderSubtle,
    },
    input: {
      flex: 1,
      color: theme.colors.textPrimary,
      fontSize: 16,
      maxHeight: 100,
      minHeight: 32,
      lineHeight: 24,
      marginRight: 8,
    },
    sendButton: {
      width: 36,
      height: 36,
      borderRadius: 18,
      backgroundColor: theme.colors.primary,
      justifyContent: 'center',
      alignItems: 'center',
    },
    sendButtonDisabled: {
      backgroundColor: `${theme.colors.primary}80`, // Add transparency for disabled
    },
    sendButtonText: {
      color: theme.colors.textPrimary,
      fontSize: 20,
      fontWeight: '700',
    },
  });

  return (
    <SafeAreaView style={themedStyles.container}>
      <StatusBar barStyle="light-content" />

      <View style={themedStyles.header}>
        <Text style={themedStyles.headerTitle}>Academi AI</Text>
        <View style={themedStyles.headerRight}>
          <TouchableOpacity
            onPress={() => setShowSuggestions(!showSuggestions)}
            style={themedStyles.suggestionsToggle}
            activeOpacity={0.7}
          >
            <Text style={themedStyles.suggestionsToggleText}>
              {showSuggestions ? 'Hide' : 'Show'} Tips
            </Text>
          </TouchableOpacity>
          <NeuralGlow isActive={isTyping} size={6} />
          <Text style={themedStyles.headerStatus}>
            {isTyping ? 'Thinking...' : 'Online'}
          </Text>
        </View>
      </View>

      <View style={themedStyles.chatContent}>
        <FlatList
          ref={flatListRef}
          data={messages}
          keyExtractor={item => item.id.toString()}
          renderItem={renderMessage}
          contentContainerStyle={themedStyles.messagesContainer}
          ListHeaderComponent={renderSuggestions}
          style={themedStyles.flatList}
          onContentSizeChange={() => {
            flatListRef.current?.scrollToEnd({ animated: true });
          }}
        />

        {renderTypingIndicator()}
      </View>

      <View style={themedStyles.inputContainer}>
        <View style={themedStyles.glassInputWrapper}>
          <TextInput
            style={themedStyles.input}
            placeholder="Ask Academi anything..."
            value={inputText}
            onChangeText={setInputText}
            placeholderTextColor={`${theme.colors.textSecondary}99`} // Add transparency
            multiline
            maxLength={500}
          />
          <TouchableOpacity
            style={[themedStyles.sendButton, inputText.trim() === '' && themedStyles.sendButtonDisabled]}
            onPress={() => sendMessage(inputText)}
            activeOpacity={0.7}
            disabled={inputText.trim() === ''}
          >
            <Text style={themedStyles.sendButtonText}>▶</Text>
          </TouchableOpacity>
        </View>
      </View>
    </SafeAreaView>
  );
}


